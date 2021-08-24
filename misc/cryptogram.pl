#!/usr/bin/env perl

use strict;
use warnings;

my $crypt = 'CGEJ, HXDDL WK QBWAQ PQ AUD KPEQA OWCGL SLWOL AW UGRD VFGJDN G EWXLN WK MWFK';
my @letters = grep { $_ ne ' ' } split '', $crypt;
my %counts;
my %subst;

$counts{$_}++ for @letters;
my @unique = sort { $counts{$b} cmp $counts{$a} } grep { /[A-Z]/ } keys %counts;

exit main();

sub disp_stats {
    for (@unique) {
        print " $_ ";
    }
    print "\n";
    for (@unique) {
        printf "%2d ", $counts{$_};
    }
    print "\n";
    for (@unique) {
        printf " %s ", ($subst{$_} || ' ');
    }
    print "\n";
}

sub disp_crypt {
    print $crypt."\n";
    for (split '', $crypt) {
        if ($_ =~ /[A-Z]/) {
            my $s = $subst{$_};
            print $subst{$_} || '_';
        } else {
            print $_;
        }
    }
    print "\n";
}

sub main {
    while(1) {
        disp_stats;

        print "\n";
        disp_crypt;

        print "Substitute XX for YY: ";
        my ($s1, $s2) = split ' ', <>;
        my @s1 = split '', $s1;
        my @s2 = split '', $s2;
        while (@s1) {
            my $s1 = uc(pop @s1);
            my $s2 = uc(pop @s2 // '');

            for (grep { $subst{$_} eq $s2 } keys %subst) {
                delete $subst{$_};
            }
            $subst{$s1} = $s2;
        }
        print "\n\n";
    }
}
