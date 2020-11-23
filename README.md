# gnames

The goal of the ``gnames`` project is to provide accurate and fast verification
of scientific names in unlimited quantities. The verification should be fast
(at least 1000 names per second) and include exact and fuzzy matching to data
from the large number of data-sources.

RESTful API of the project is described by the [OpenAPI Specification]

## Known limitations that we want to address in the future

- Exact matches of a misspelling to a name-string in a bad quality database
prevent to find fuzzy matches.

To increase performance we stop any further tries if name matched succesfully.
This prevents finding how mispelling would fuzzy-match name-strings in better
quality databases.

- Fuzzy matching of a name where genus string is broken by a space.

For example we cannot match 'Abro stola triplasia' to 'Abrostola triplasia'.
There is only 1 edit distance between the strings, however we stem specific
epithets, so in reality we fuzzy-match 'Abro stol triplas' to 'Abrostola triplas'.
That means now we have edit distance 2 which is usually beyond our threshold.
See issue `https://github.com/gnames/gnmatcher/issues/19`

[OpenAPI Specification]  https://app.swaggerhub.com/apis/dimus/gnames/1.0.2