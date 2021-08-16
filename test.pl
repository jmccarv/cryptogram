#!/usr/bin/env perl
#
use strict;
use warnings;

my @words;
while (<>) {
    last unless /^(\S+)\s+(\d+)/;
    $words[length($1)]->{$1} = $2;
}

my $w = $words[3];
for (sort { $w->{$a} <=> $w->{$b} } keys %$w) {
    print "$_ $w->{$_}\n";
}
