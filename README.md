# go
General purpose libraries / APIs for use in mlab code.

## General guidance
Packages in this repo should be:
+ Useful across multiple other packages
+ Non-trivial, either in lines of code or in semantic complexity.
Small simple things should likely just be defined where they are used.
+ Fairly carefully designed.  Probably should review design with other
engineers before putting in too much effort.

Note that packages here are intended to be used in *other* repositories.
This means that it will be somewhat disruptive to change APIs in these
packages, as API changing PRs will break other repos, and require additional
PRs to fix those repositories.

## package tags
Please mark packages in development as *alpha* or *beta*.  Use of these packages
should be discouraged in other repositories, until they are *stable*.

Once a packages API has stabilized, mark the package as *stable*.

You can still add __new__ APIs to stable packages, but mark these new APIs
as *alpha* or *beta* until they are regarded as stable and suitable for
general use.

## packages
### cloudtest
Utilities for testing google cloud service abstractions.

### bqutil

