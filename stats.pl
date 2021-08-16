#!/usr/bin/env perl

use strict;
use warnings;

my %lens;
my %letters;
my $tot;
my $wc;
while (<>) {
    chomp;
    next unless /^[a-z]+$/i;
    $wc++;
    $lens{length($_)}++;
    $letters{uc($_)}++ for (split '', $_);
    $tot += length($_);
}

print "$wc Words Counted\n\n";
print "Count  Word Length\n";
for (sort { $lens{$a} <=> $lens{$b} } keys %lens) {
    printf "%5d  %2d\n", $lens{$_}, $_;
}

print "\nLetter Frequency Percentage\n";
my $pct;
for (sort { $letters{$b} <=> $letters{$a} } keys %letters) {
    print "$_  ";
    $pct .= sprintf("%-2.0f ", ($letters{$_}/$tot*100));
}
print "\n$pct\n";
