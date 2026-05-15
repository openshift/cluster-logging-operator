#!/usr/bin/env perl
use strict;
use warnings;

my $file = $ARGV[0];
open(my $fh, '<', $file) or die "Cannot open $file: $!";
my @lines = <$fh>;
close($fh);

# Strip carriage returns
s/\r//g for @lines;

# Escape || so AsciiDoc does not treat them as column spans
s/\|\|/\\|\\|/g for @lines;

# Wrap bare {…} in backticks so AsciiDoc does not treat them as attribute refs
s/(\{[^}]*\})/`$1`/g for @lines;

# Join multi-line table cells onto a single line
my @result;
my $in_table = 0;
my $i = 0;

while ($i < @lines) {
    my $line = $lines[$i];

    if ($line =~ /^\|\s*=/) {
        $in_table = !$in_table;
        push @result, $line;
        $i++;
        next;
    }

    if ($in_table && $line =~ /^\|/) {
        my $row = $line;
        chomp($row);

        while ($i + 1 < @lines && $lines[$i + 1] !~ /^(\||\[|=)/) {
            $i++;
            my $cont = $lines[$i];
            next if $cont =~ /^\s*$/;
            chomp($cont);
            $cont =~ s/^\s+//;
            $row .= ' ' if $row !~ /[\s|]$/ && $cont !~ /^\s/;
            $row .= $cont;
        }
        $row .= "\n";
        push @result, $row;
        $i++;
        next;
    }

    push @result, $line;
    $i++;
}

# Escape stray pipe characters inside table description cells
foreach my $line (@result) {
    next unless $line =~ /^\|/ && $line !~ /^\|\s*=/;

    $line =~ s/(\d)-(\d+)([A-Z])/$1-$2 $3/g;

    if ($line =~ /^(\|[^|]*\|[^|]*\|)(.*)$/) {
        my ($prefix, $desc) = ($1, $2);
        $desc =~ s/\|/\\|/g;
        $line = $prefix . $desc . "\n";
    }
}

open($fh, '>', $file) or die "Cannot write to $file: $!";
print $fh @result;
close($fh);
