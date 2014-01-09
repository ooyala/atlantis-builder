#!/usr/bin/env perl
package Builder;

use strict;
use warnings;

use Cwd qw(realpath);
use Data::Dumper;
use File::Basename;
use File::Copy qw(copy);
use File::Path;
use File::Temp;

use Carp::Always;

my $home = dirname(realpath($0));

# Save script start time for reports in case of failure.
my $started = localtime;

# Read standard config
my $config_file = "builder-config.toml";
my ($config, $error) = TOML->from_toml(scalar slurp($config_file));
unless ($config) {
	die "error parsing $config_file: $error"
}

# Default configuration for error notification emails.
my $enable_email = $ENV{'EMAIL_ON_ERROR'} ? 1 : undef;
my $team_address = $config->{'owner_address'};
my $maintainer = $config->{'maintainer'};

sub email_team {
	return unless $enable_email;
	my($line, @message) = @_;

	open(my $sendmail, "| sendmail -t");
	print $sendmail "From: $config->{'from_address'}" . "\n";
	print $sendmail "To: $config->{'owner_address'}" . "\n";
	print $sendmail "Cc: $team_address" . "\n";
	print $sendmail 'Subject: builder.pl failed' . "\n";
	print $sendmail 'Content-Type: text/plain' . "\n";
	print $sendmail "\n";
	print $sendmail <<BODY;
		invoked... $started
		as........ $0 $ARGV[0] $ARGV[1] $ARGV[2]
		line...... $line
		message... @message
BODY
	close($sendmail);
}

sub slurp {
	my $file = shift;

	# If you do not understand the next line, RUN! Don't look back.
	local $/;

	open(my $fh, '<:encoding(UTF-8)', $file) || die "$file - $!";
	my $data = <$fh>;
	close($fh);

	return $data;
}

sub spew {
	my $file = shift;
	my $data = join("\n", @_);

	open(my $fh, '>:encoding(UTF-8)', $file) || die "$file - $!";
	print $fh $data;
	close($fh);

}

sub read_manifest {
	my $dir2build = shift;

	my $toml_file =  $dir2build . "/manifest.toml";
	copy($toml_file, "$home/manifest.toml");

	my ($manifest, $error) = TOML->from_toml(scalar slurp($toml_file));
	unless ($manifest) {
		die "error parsing $toml_file: $error"
	}
	print "manifest\n" . Dumper($manifest);
	return $manifest;
}

sub run_commands {
	foreach (@_) {
		my $command = $_;
		$command =~ s/^\s+//;
		$command =~ s/\s+$//;
		print "$command\n";
		system($command) == 0 || die "$command";
	}
}

# File::Temp->newdir() removes the temporary directory when the script exits.
sub git_checkout {
	my $git_url = shift;
	my $git_sha = shift;
	my $git_dir = File::Temp->newdir("$home/" . basename($git_url) . "-XXXX");

	run_commands(
		"git clone $git_url $git_dir",
		"cd $git_dir && git checkout $git_sha",
		"cd $git_dir && git submodule update --init",
	);

	return $git_dir;
};

sub docker_exists {
	my $exists = undef;

	my $container = $ENV{'REGISTRY_HOST'} . "/apps/" . join('-', @_);

	my $command = "sudo docker images";
	open(my $docker, "$command |") || die "$command - $!";
	while (my $line = <$docker>) {
		if (grep /^$container\>/, $line) {
			$exists = 1;
		}
	}
	close($docker);

	unless ($exists) {
		# FIXME(manas): How do we tell if docker pull failed because it did not find
		# the container, or some out-of-band error occurred?
		if (system("sudo docker pull $container") == 0) {
			$exists = 1;
		}
	}

	return $exists;
}

sub make_app_dir {
	my ($manifest, $dir2build) = @_;

	my $app_dir = File::Temp->newdir("$home/$manifest->{'name'}" . "-XXXX");
	mkpath("$app_dir/build/info");
	mkpath("$app_dir/build/app");

	run_commands(
		"echo $started > $app_dir/build/info/time",
		"echo $ARGV[1] > $app_dir/build/info/branch",
		"cd $dir2build && git rev-list --all > $app_dir/build/info/revlist",
		"cd $dir2build && rsync -av --exclude .git --exclude manifest.toml ./* $app_dir/build/app/"
	);

	return $app_dir;
}

sub setup_runit {
	my ($app_dir, $manifest) = @_;

	my $template_dir = $home . "/templates";
	my $app_template = $home . "/templates/runit_app_run";

	mkpath("$app_dir/etc/sv/rsyslog");
	spew("$app_dir/etc/sv/rsyslog/run", "#!/bin/bash", "exec rsyslogd -nc5");
	chmod 755, "$app_dir/etc/sv/rsyslog/run";

	# Dynamic array where each line of the container rsyslog config will be added as generated
	# to be written out after completely created.
	my @container_conf;

	my $cmdN = 0;
	# This makes sure we handle both strings and arrays of strings
	foreach ( ref($manifest->{'run_command'}) ? @{$manifest->{'run_command'}} : $manifest->{'run_command'} ){
		my $run_cmd = $_;

		my $run_dir = "$app_dir/build/runit$cmdN";
		mkpath("$run_dir");

		my $app_run_file = slurp($app_template);
		$app_run_file =~ s#__RUN_COMMAND__#$run_cmd#;
		$app_run_file =~ s#__APP_FACILITY__#local$cmdN#g;
		spew("$run_dir/run", $app_run_file);
		chmod 0755, "$run_dir/run";

		# Write out rsyslog config
		my $syslog_dir = "/var/log/atlantis/syslog/app${cmdN}";
		my $m10 = 10*1024*1024;
		my $logrot = "/etc/atlantis/logrot";
		push @container_conf, qq(\$outchannel App${cmdN}Info,${syslog_dir}/info.log,${m10},${logrot});
		push @container_conf, qq(\$outchannel App${cmdN}Error,${syslog_dir}/error.log,${m10},${logrot});
		push @container_conf, qq(\$outchannel App${cmdN}All,${syslog_dir}/all.log,${m10},${logrot});
		push @container_conf, "";
		push @container_conf, "local${cmdN}.=info -?App${cmdN}Info";
		push @container_conf, "local${cmdN}.=error -?App${cmdN}Error";
		push @container_conf, "local${cmdN}.* -?App${cmdN}All";
		push @container_conf, "";

		$cmdN += 1;
	}

	mkpath("$app_dir/etc/rsyslog.d");
	spew("$app_dir/etc/rsyslog.d/01-container.conf", @container_conf);
}

sub create_dockerfile {
	my ($app_dir, $manifest) = @_;

	my $dockerfile = slurp("$home/templates/$manifest->{'app_type'}/Dockerfile");
	$dockerfile =~ s#__REGISTRY_HOST__#$ENV{'REGISTRY_HOST'}#;
	$dockerfile =~ s#__MAINTAINER__#$maintainer#;
	$dockerfile =~ s#__IMAGE__#$manifest->{'image'}#;
	$manifest->{'setup_command'} = "echo No setup!" unless $manifest->{'setup_command'};
	$dockerfile =~ s#__SETUP_COMMAND__#$manifest->{'setup_command'}#;
	spew("$app_dir/Dockerfile", $dockerfile);

	return "$app_dir/Dockerfile";
}

# Trap the __DIE__ handler, so we can email error reports if the build fails.
$SIG{__DIE__} = sub {
	my @locations = caller();
	email_team($locations[2], @_);
	print "[die] @_\n";
	exit(-1);
};

unless ($#ARGV == 2) {
	die "usage: builder.pl <git_url> <git_sha> <root_dir>\n";
}

unless ($ENV{'REGISTRY_HOST'}) {
	die "REGISTRY_HOST not in environment!"
}

my $git_dir = git_checkout(@ARGV);
my $dir2build = $git_dir . $ARGV[2];

my $manifest = read_manifest($dir2build);
if ($manifest->{'enable_email'}) {
	$enable_email = 1;
	$team_address = $manifest->{'email'};
}

if (docker_exists($manifest->{'name'}, $ARGV[1])) {
	unless ($ENV{'REBUILD_CONTAINER'}) {
		print "image for $manifest->{'name'}-$ARGV[1] exists\n";
		exit(0);
	} else {
		print "rebuilding $manifest->{'name'}-$ARGV[1]\n";
	}
}

my $app_dir = make_app_dir($manifest, $dir2build);

setup_runit($app_dir, $manifest);

my $prebuild = "$home/templates/$manifest->{'app_type'}/prebuild";
if (-e $prebuild) {
	copy($prebuild, "$app_dir/prebuild");
	run_commands(
		"chmod +x $app_dir/prebuild",
		"cd $app_dir && ./prebuild"
	);
}

create_dockerfile($app_dir, $manifest);

# Run the docker build and push a tag if the build succeeds.
my $container = "$ENV{'REGISTRY_HOST'}/apps/$manifest->{'name'}-$ARGV[1]";
run_commands(
	"sudo docker build -rm -t \"$container\" $app_dir",
	"sudo docker push $container"
);

# -------------------------------------------------------------------
# TOML - Parser for Tom's Obvious, Minimal Language.
#
# Copyright (C) 2013 Darren Chamberlain <darren@cpan.org>
# -------------------------------------------------------------------

package TOML;

use strict;
use warnings;

use B;
use Text::Balanced qw(extract_bracketed);

my %UNESCAPE = (
	q{b}  => "\x08",
	q{t}  => "\x09",
	q{n}  => "\x0a",
	q{f}  => "\x0c",
	q{r}  => "\x0d",
	q{"}  => "\x22",
	q{/}  => "\x2f",
	q{\\} => "\x5c",
);

sub from_toml {
	my $class = shift;
	my $string = shift;
	my %toml;   # Final data structure
	my $cur;
	my $err;    # Error
	my $lineno = 0;

	# Normalize
	$string =
	join "\n",
	grep !/^$/,
	map { s/^\s*//; s/\s*$//; $_ }
	map { s/#.*//; $_ }
	split /[\n\r]/, $string;

	while ($string) {
		# strip leading whitespace, including newlines
		$string =~ s/^\s*//s;
		$lineno++;

		# Store current value, to check for invalid syntax
		my $string_start = $string;

		# Strings
		if ($string =~ s/^(\S+)\s*=\s*"(.+)"\s*//) {
			my $key = "$1";
			my $val = "$2";
			$val =~ s/^"//;
			$val =~ s/"$//;
			$val =~ s!
			\\([btnfr"/\\])
			|
			\\u([0-9A-Fa-f]{4})
			!
			if (defined $1) {
			$UNESCAPE{$1}
			} else {
			pack "U", hex $2;
			}
			!gex;

			if ($cur) {
				$cur->{ $key } = $val;
			}
			else {
				$toml{ $key } = $val;
			}
		}

		# Boolean
		if ($string =~ s/^(\S+)\s*=\s*(true|false)//i) {
			my $key = "$1";
			my $num = lc($2) eq "true" ? "true" : "false";
			if ($cur) {
				$cur->{ $key } = $num;
			}
			else {
				$toml{ $key } = $num;
			}
		}

		# Date
		if ($string =~ s/^(\S+)\s*=\s*(\d\d\d\d\-\d\d\-\d\dT\d\d:\d\d:\d\dZ)\s*//) {
			my $key = "$1";
			my $date = "$2";
			if ($cur) {
				$cur->{ $key } = $date;
			}
			else {
				$toml{ $key } = $date;
			}
		}

		# Numbers
		if ($string =~ s/^(\S+)\s*=\s*([+-]?[\d.]+)(?:\n|\z)//) {
			my $key = "$1";
			my $num = $2;
			if ($cur) {
				$cur->{ $key } = $num;
			}
			else {
				$toml{ $key } = $num;
			}
		}

		# Arrays
		if ($string =~ s/^(\S+)\s=\s*(\[)/[/) {
			my $key = "$1";
			my $match;
			($match, $string) = extract_bracketed($string, "[]");
			if ($cur) {
				$cur->{ $key } = eval $match || $match;
			}
			else {
				$toml{ $key } = eval $match || $match;
			}
		}

		# New section
		elsif ($string =~ s/^\[([^]]+)\]\s*//) {
			my $section = "$1";
			$cur = undef;
			my @bits = split /\./, $section;

			for my $bit (@bits) {
				if ($cur) {
					$cur->{ $bit } ||= { };
					$cur = $cur->{ $bit };
				}
				else {
					$toml{ $bit } ||= { };
					$cur = $toml{ $bit };
				}
			}
		}

		if ($string eq $string_start) {
			# If $string hasn't been modified by this point, then
			# it contains invalid syntax.
			(my $err_bits = $string) =~ s/(.+?)\n.*//s;
			return wantarray ? (undef, "Syntax error at line $lineno: $err_bits") : undef;
		}
	}

	return wantarray ? (\%toml, $err) : \%toml;
}
