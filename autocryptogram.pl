#!/usr/bin/env perl

use strict;
use warnings;

my $crypt = 'CGEJ, HXDDL WK QBWAQ PQ AUD KPEQA OWCGL SLWOL AW UGRD VFGJDN G EWXLN WK MWFK';
my $dict = 'aspell_dump_master.out';
my %words;

exit main();

# divide word list into multiple lists, one for each word length
# 

sub main {
    open(my $fh, "<", $dict) or die "Failed to open '$dict': $!";
    while (<$fh>) {
        chomp;
        $_ = uc($_);
        next unless /^[A-Z]+$/;

        push @{$words{length($_)}}, $_;
    }
}
