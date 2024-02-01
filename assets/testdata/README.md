# PKZip for DOS Tests

This is a collection of text and binaries compressed with the various first editions of PKWare's PKZIP for DOS. Each major revision introduced new compression methods that were downwardly but not upwardly compatible. This meant PKZIP 2.0 could unpack archives created using the PKZIP 1.x Implode method, but couldn't create new archives using that method.

Special thanks to [Ben Garrett](https://github.com/bengarrett) for creating these test files.

* `PKZ80*.ZIP` were compressed using PKZip 0.8 for DOS.
* `PKZ90*.ZIP` with 0.90.
* `PZ110*.ZIP` with 1.10.
* `PKZIP204*.ZIP` with 2.04.

## Reduced method

Those files names with the prefix `A[1-4]` used the Reduced ASCII method and the numbers signify their compression factor from least to more. `B[1-4]` use the Reduced binary method.

## Implode and Shrink methods

* `PKZIP110EI.ZIP` uses the *Implode* method only.
* `PKZIP110ES.ZIP` uses the *Shrink* method only.
* `PKZIP110EX.ZIP` uses *maXimal compression*.
* `PKZIP110.ZIP` was compressed without any parameters provided which should use a combination of methods.

## Deflate method

* `PKZIP204EX.ZIP` uses *extra* compression with the deflate method.
* `PKZIP204EN.ZIP` uses *normal (default)* compression.
* `PKZIP204EF.ZIP` uses *fast*.
* `PKZIP204ES.ZIP` uses *super fast*.
* `PKZIP204E0.ZIP` uses *no compression*.
