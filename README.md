#                       So You Want To Solve Cryptograms
#                                      or
#                        There's Got To Be A Better Way

Let's start with what a cryptogram is. A cryptogram is a simple encryption
scheme where each letter in the alphabet is substituted for another. This
substitution remains constant for the entire message. Typically cryptograms
never map a letter to itself.

So for exmample using the following substitution key:
```
encoded  ABCDEFGHIJKLMNOPQRSTUVWXYZ
decoded  TO  I  P CM DE G NRAS FUH
```

SCIENCE IS THE GREAT ANTIDOTE TO THE POISON OF ENTHUSIASM AND SUPERSTITION.

encodes to

VKFOSKO FV BZO QTOUB USBFNCBO BC BZO ICFVCS CX OSBZYVFUVL USN VYIOTVBFBFCS.

------

## This approach

We are trying to find the substitution key used to encrypt the message.
We start with an empty key where each letter maps to nothing (null):

```
  from: ABCDEFGHIJKLMNOPQRSTUVWXYZ
    to: __________________________
```

We have a list of known words (English in our case).
Take one of the encrypted words and choose a word from our word
list to try as the decrypted word. Now assuming that word is the
correct decryption, we try to populate our solution key using that
word. If we can't populate it because any of the letters have already
been used by our key then we try another word from our list.

For example, using the above cryptogram, we might choose to try
'THOUGHT' for the first encrypted word 'VKFOSKO'. In that case we
would populate our solution key like so.

```
  T
  from: ABCDEFGHIJKLMNOPQRSTUVWXYZ
    to: _____________________T____
 
  H
  from: ABCDEFGHIJKLMNOPQRSTUVWXYZ
    to: __________H__________T____

  O
  from: ABCDEFGHIJKLMNOPQRSTUVWXYZ
    to: _____O____H__________T____

  U
  from: ABCDEFGHIJKLMNOPQRSTUVWXYZ
    to: _____O____H___U______T____

  G
  from: ABCDEFGHIJKLMNOPQRSTUVWXYZ
    to: _____O____H___U___G__T____

  H
  from: ABCDEFGHIJKLMNOPQRSTUVWXYZ
    to: _____O____H___U___G__T____
```

Now we get to the final 'T' and we have a problem, In our original encrypted
word, 'VKFOSKO', the letter in this position is 'O', which we've already
decided would have decode to 'U' if the decrypted word were 'THOUGHT'. So
now we know that 'THOUGHT' cannot be the correct decrypted word.
In this case we now choose another word to try. Read on to see how we avoid
ever trying a word like this that cannot be correct.

Maybe the next word we try is 'INSTANT', which will fit. Then our key is:

```
  from: ABCDEFGHIJKLMNOPQRSTUVWXYZ
    to: _____S____N___T___A__I____
```

Since we successfully found a word, we choose another encrypted word and
try the same process again. Maybe we try the word 'INSTANT' for the
encrypted word 'USBFNCBO'

```
  I
  from: ABCDEFGHIJKLMNOPQRSTUVWXYZ
    to: _____O____H___U___G_IT____
```

When we try the 'N' we find that the encrypted letter 'S' already has a
mapping in our key, to 'G'. So 'INSTANT' cannot fit in our current solution.
So we can move on to the next word to try for this encrypted word.

We continue on in this manner until we've worked our way through all the
encrypted words and found a possible solution. That solution is the scored
(see below). At the end of the program run (or during if -p is specified)
the highest scored solutions will be displayed.

If we work through our word list and never find a word that fits for
a given encrypted word, that's considered an 'unknown' word. By default
the program will not allow any unknown words in a solution. You can
change this with -u and allow (and indeed check for) unknown words in
your solution. It will slow down the program but you may get better results.



This solver relies on a word frequency dictionary. This is in the freq.txt
file in this distribution and is a simple list of words, one per line,
followed by the number of times that word shows up in English literature.

We start with an empty solution key (SK), which we will attempt to populate
as we go.

The cryptogram text is broken down into a list of unique words. 
Even though the word is encrypted, it still has useful properties for us:
  
  * The word length
  * The pattern of the letters in the word

Word lengths and patterns are useful in helping narrow our search for
possible decodings of an encrypted word. Consider the following word:

```
  MISSISSIPPI
```

We can calculate a pattern of letters for this word by assigning a
number to each unique letter in the word, like so:

```
  MISSISSIPPI
  12332332442
```

Any encrypted version of this word will have the same length and pattern.
We can use this to our advantage to only consider words in our word list
with the same length and pattern. In fact, we only need to consider the
pattern, since the length affects the pattern of a word!

Now we start by looking at each unique encrypted word. For each encrypted
word (EW) we consider each possible word from our dictionary (DW).
We try to populate our solution key, but if we've already assigned a
mapping between an encrypted letter and a clear text letter we know
this word will not work with our current mapping and so move on to the
next word.

More to come...
