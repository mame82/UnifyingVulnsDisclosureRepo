# Observed report sequences

K -> encrypted key report
L -> LED report
other reports ignored

Note: An encrypted keyboard report frame followed by "set keep alive" frames is a strong indicator for a
key release report, but this kind of distinguishing key down vs key release isn't needed if LED output
reports are available

## K360, Linux

KLK --> key,led,key: first K is key down, second is key up (LED on togle, after key down)
KKL --> key,key,led: first K is key down, second is key up (LED off toggle, after key up)

## K360, Windows

KL*K --> key,led,key: first K is key down, second is key up (LED on and off toggle, after key down; LED report repeated if key pressed long)


## K400+, Linux

KL*K --> key,led,key: first K is key down, second is key up (LED on and off toggle, after key down; LED report repeated if key pressed long)

## K400+, Windows (frequent Keep alives between LED reports, thus same as K360 on Linux)

KLK --> key,led,key: first K is key down, second is key up (LED on togle, after key down)
KKL --> key,key,led: first K is key down, second is key up (LED off toggle, after key up)

## Conclusion

K400+ vs K360 have exact opposite behavior on Linux vs Windows

Repeated LED reports could be eliminated, by ignoring LED reports with same content (no LED toggle)

Comparing successive LED reports, allows to determin the LED which toggled and thus the key (CAPS, SCROLL or NUM)
which was encoded in key down report. For key up reports all keys are 0x00. We don't take care of possible modifiers.

So a kee-down-key-release sequence is either KKL or KLK if a LED is toggled.

We need about 13 of those sequences with **successive counters** for every K, in order to have enough known plaintext
to eliminate XOR encoded key presses (whitening of reports).

The LED approach does not need to account for "set keep alive", which is less reliable (doesn't occur on all key releases).

