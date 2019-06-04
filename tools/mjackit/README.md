tbd

Please read "report 1" carefully. As you may notice, the report lacks two things, which are needed 
in order to reproduce everything:

1) The AES input data blob, needed for encryption/decryption
2) The pseudo-code for the substitution algorithm of the keys

Both missing parts are part of `mjackit` implementation. The final notes of all my reports contain this 
statement:

```
Valid key material won’t be disclosed at any point, as it doesn’t add up to improve security or to 
protect end users
```

I do not consider the AES input data or substitution algorithm to be `valid key material`. Both information
belong to the root cause of the security issue described in report 1. The issue described in this report 
results in generation of weak key material, though ()which could be reproduced using the code from `mjackit`).

Because of this, I want to hear your opinion on disclosing this content, before I add in the source code for 
`mjackit`. 