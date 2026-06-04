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

# Escape bare {…} with passthrough so AsciiDoc does not treat them as attribute refs
s/(\{[^}]*\})/pass:[$1]/g for @lines;

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

        my $has_list = 0;
        my @cont_lines;
        while ($i + 1 < @lines && $lines[$i + 1] !~ /^(\||\[|=)/) {
            $i++;
            my $cont = $lines[$i];
            next if $cont =~ /^\s*$/;
            chomp($cont);
            $cont =~ s/^\s+//;
            if ($cont =~ /^\d+\.\s|^[-*]\s/) {
                $has_list = 1;
            }
            unless ($cont =~ /^a\|/) {
                $cont =~ s/(?<!\\)\|/\\|/g;
            }
            push @cont_lines, $cont;
        }

        if ($has_list) {
            # Use a| cell for multi-line content with lists;
            # blank line before list and after content closes the a| cell
            $row =~ s/^(\|[^|]*\|[^|]*)\|/$1\na| /;
            my $prev_was_list = 0;
            foreach my $cont (@cont_lines) {
                my $is_list = ($cont =~ /^\d+\.\s|^[-*]\s/);
                if ($is_list && !$prev_was_list) {
                    $row .= "\n";
                }
                if ($is_list) {
                    $cont =~ s/^\d+\.\s/. /;
                }
                $row .= "\n" . $cont;
                $prev_was_list = $is_list;
            }
            $row .= "\n\n";
        } else {
            foreach my $cont (@cont_lines) {
                $row .= ' ' if $row !~ /[\s|]$/ && $cont !~ /^\s/;
                $row .= $cont;
            }
            $row .= "\n";
        }
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
        # Escape only literal pipes inside the description text,
        # but keep real column delimiters (" | ") intact.
        $desc =~ s/(?<!\\)\|(?!\s)/\\|/g;
        $line = $prefix . $desc . "\n";
    }
}

open($fh, '>', $file) or die "Cannot write to $file: $!";
print $fh @result;
close($fh);
