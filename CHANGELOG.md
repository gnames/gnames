# Changelog

The GNames project follows [Semantic Versioning guidelines].

## Unreleased

## [v0.8.0] - 2022-02-21 Mon

- Add [#86]: make Input.DataSources and Input.WithAllMatches behave
  similar for Verification and Search. DataSources now limit search
  to provided data-sources, while WithAllMatches shows all Results.
  There is no option anymore to provide Results 1 per data-source
  anymore.
  WARNING: Introduces backward incompatibility and changes in
  https://apidoc.globalnames.org/gnames-beta documentation.

- Fix [#84]: MatchType for `Jsoetes longissimum` is `NoMatch` instead of
  `Fuzzy`. Show `Isoetes longissimum` and `Isoetes longissima` in results.
- Fix [#85]: search with `tx:` works correctly.

## [v0.7.1] - 2022-02-14 Mon

- Add: use /api/v0 instead of /vpi/v1 for gnmatcher.

## [v0.7.0] - 2022-02-14 Mon

- Add [#82]: exact match by stemed canonical can return alternative suffixes.

## [v0.6.5] - 2022-02-09 Wed

- Fix: add missing env names for NSQ filters.

## [v0.6.4] - 2022-02-09 Wed

- Add [#81]: optional filters for NSQ logs.

## [v0.6.3] - 2022-02-08

- Add [#80]: more information for NSQ logs. Change logs lib to zerolog.
- Fix [#79]: OverloadDetected is not created for names without overload.

## [v0.6.2] - 2022-02-06

- Fix: limit NSQ to queries with `verifications` and `search` only.

## [v0.6.1] - 2022-02-06

- Add: improve README.
- Fix: typos in OverloadDetected values.

## [v0.6.0] - 2022-02-06

- Add update changelog to follow older changelog convention to avoid
  Markdown linting problems and to make changelog more readable.
- Add [#78]: optional NSQd log aggregation.
- Add [#77]: verification of virus names.
- Add OverloadDetected field in the output to indicate when
  there are too many results returned from the database.

## [v0.5.12] - 2022-01-30

- Add: allow searches with one-letter authors
  (related to `https://github.com/gnames/gnverifier/issues/75`).

## [v0.5.11] - 2021-12-15

- Add: make name field `n:` for faceted searc more powerful.

## [v0.5.10] - 2021-12-09

- Add: update dependency modules.

## [v0.5.9] - 2021-12-08

- Fix: small changes in scoring algorithm for authors.

## [v0.5.8] - 2021-12-07

- Add [#75]: add score details to JSON output.

## [v0.5.7] - 2021-12-06

- Add: dereferenced slice of names in verification.Output.

## [v0.5.6] - 2021-12-05

- Fix: fixes in context/kingdoms stats

## [v0.5.5] - 2021-12-04

- Fix: fixes in verificaton GET parameters.

## [v0.5.4] - 2021-12-04

- Add: modifications in input/output format.

## [v0.5.3] - 2021-12-04

- Add: modifications in input/output formats.

## [v0.5.2] - 2021-12-03

- Add [#73]: allow multiple data sources.

## [v0.5.1] - 2021-12-02

- Add: improve output, bug fixes.

## [v0.5.0] - 2021-12-01

- Add [#71]: faceted search API.

## [v0.4.0] - 2021-11-22

- Add [#51]: metadata to verification, optional context calculation [#51].
  WARNING: Introduces major backward incompatibility

## [v0.3.3] - 2021-10-30

- Fix: missing GET parameter added: `all_matches`.

## [v0.3.2] - 2021-10-28

- Add [#70]: all sources/all matches results are sorted by score.

## [v0.3.1] - 2021-10-25

- Fix: remove dependency on `dgrijalva/jwt-go` which has
  security problems.

## [v0.3.0] - 2021-10-25

- Add [#69]: optional parameters for returning all matched sources,
  all matched results.

## [v0.2.2] - 2021-10-24

- Fix [#68]: data sources show correctly.

## [v0.2.1] - 2021-10-22

- Add: sort data-sources by ID.

## [v0.2.0] - 2021-04-09

- Add [#67]: add an option to capitalize the first letter of a
  names-string.
- Add [#61]: ClassificationIDs to verification output.
- Add: update gnparser to v0.2.0.
- Add: update gnmatcher to v0.5.7.

## [v0.1.8] - 2021-03-22

- Add: update gnparser to v1.1.0.
- Add [#65]: stop processing canceled, or expired requests.
- Add: update dependenies to gnparser v1.0.5, gnmatcher v0.5.5.

## [v0.1.7] - 2021-01-24

- Add [#62]: update gnparser to v1.0.4.

## [v0.1.6] - 2020-12-15

- Add [#59]: the score entity is now based on an interface.
- Add [#56]: reduce impact of names with huge number of instances that
  slowed down name verification significantly.
- Add [#54]: Improve documentation.

## [v0.1.5] - 2020-12-09

- Add: Change server timeout for reading and writing to 5 min.

## [v0.1.4] - 2020-12-08

- Add [#53]: Introduce GET method to make ad-hoc verifications easier.
- Add [#52]: Score calculation uses parse quality. That helps push names
  that parsed better to the top.

## [v0.1.3] - 2020-12-07

- Add [#47]: new field `isOutlinkReady` for DataSources. This field marks
  data-sources that are prepared to be used as outlinks (for example at BHL).
- Add [#46]: outlink URLs are now provided in the results.
- Add [#45]: DataSource output cleaned up.

## [v0.1.2] - 2020-12-02

- Add: set the gnames web service at `https://verifier.globalnames.org`.
- Add: set dockerhub for releases.

## [v0.1.1] - 2020-12-02

- Update depenency to `gnlib`.

## [v0.1.0] - 2020-11-22

- Add [#42]: improve architecture, add OpenAPI.
- Add [#41]: make the code compatible with gnmatcher v0.3.6.
- Add [#40]: refactor entities, move some of them to gnlib.

## [v0.0.4] - 2020-10-25

- Add [#35]: increase priority for the authors score.
- Fix [#38]: 'Acacia vestita may' matches with `PartialExact` to
  'Acacia vestita'. Now to register Fuzzy there is a limit of 5 characters
  in a word per edit distan event.
- Fix [#36] score calculation uses `edit distance` correctly for fuzzy matches.
- Fix [#33]: provide processing of unparseable accepted names.
- Fix [#32]: set false positive from gnmatcher as NoMatch. Bloom filters
  create rare false positives. Check every returned name for correctness
  using Levenshtein automata.

## [v0.0.3] - 2020-09-16

- Add [#24]: get reasonable preferred matches from the real data.
- Add [#28]: reasonable BestMatch from the real data.
- Add [#19]: decrease score for higher edit distance fuzzy matching.
- Add [#20]: currently accepted names generate higher score than synonyms.
- Add [#15]: use curation level for scoring results.
- Add [#16]: use authorship for score calculation.
- Add [#17]: use infrapecific ranks for score calculation.
- Add [#18]: develop a ranking system for score calculation.
- Add [#23]: make gnmatcher functionality interface-based. This allows to
  choose to use gnmatcher as a service or as a library in the future.

## [v0.0.2] - 2020-09-11

- Add [#21]: improve the code architecture using 'clean architecture'
  principles.

## [v0.0.1] - 2020-09-05

- Add [#14]: return complete result.
- Add [#13]: get DataSource metadata out of API.
- Add [#8]: make decode/encode accept either Gob or JSON.
- Add [#6]: migrate from protobuf to Go.
- Add [#5]: incorporate gnmatcher service.
- Add [#4]: send names via HTTP API.
- Add [#3]: setup testing framework.
- Add [#1]: develop a draft of output format as a protobuffer.

## [v0.0.0] - 2020-05-25

- Add initial commit

## Footnotes

This document follows [changelog guidelines]

[v0.8.0]: https://github.com/gnames/gnames/compare/v0.7.1...v0.8.0
[v0.7.1]: https://github.com/gnames/gnames/compare/v0.7.0...v0.7.1
[v0.7.0]: https://github.com/gnames/gnames/compare/v0.6.5...v0.7.0
[v0.6.5]: https://github.com/gnames/gnames/compare/v0.6.4...v0.6.5
[v0.6.4]: https://github.com/gnames/gnames/compare/v0.6.3...v0.6.4
[v0.6.3]: https://github.com/gnames/gnames/compare/v0.6.2...v0.6.3
[v0.6.2]: https://github.com/gnames/gnames/compare/v0.6.1...v0.6.2
[v0.6.1]: https://github.com/gnames/gnames/compare/v0.6.0...v0.6.1
[v0.6.0]: https://github.com/gnames/gnames/compare/v0.5.12...v0.6.0
[v0.5.12]: https://github.com/gnames/gnames/compare/v0.5.11...v0.5.12
[v0.5.11]: https://github.com/gnames/gnames/compare/v0.5.10...v0.5.11
[v0.5.10]: https://github.com/gnames/gnames/compare/v0.5.9...v0.5.10
[v0.5.9]: https://github.com/gnames/gnames/compare/v0.5.8...v0.5.9
[v0.5.8]: https://github.com/gnames/gnames/compare/v0.5.7...v0.5.8
[v0.5.7]: https://github.com/gnames/gnames/compare/v0.5.6...v0.5.7
[v0.5.6]: https://github.com/gnames/gnames/compare/v0.5.5...v0.5.6
[v0.5.5]: https://github.com/gnames/gnames/compare/v0.5.4...v0.5.5
[v0.5.4]: https://github.com/gnames/gnames/compare/v0.5.3...v0.5.4
[v0.5.3]: https://github.com/gnames/gnames/compare/v0.5.2...v0.5.3
[v0.5.2]: https://github.com/gnames/gnames/compare/v0.5.1...v0.5.2
[v0.5.1]: https://github.com/gnames/gnames/compare/v0.5.0...v0.5.1
[v0.5.0]: https://github.com/gnames/gnames/compare/v0.4.0...v0.5.0
[v0.4.0]: https://github.com/gnames/gnames/compare/v0.3.3...v0.4.0
[v0.3.3]: https://github.com/gnames/gnames/compare/v0.3.2...v0.3.3
[v0.3.2]: https://github.com/gnames/gnames/compare/v0.3.1...v0.3.2
[v0.3.1]: https://github.com/gnames/gnames/compare/v0.3.0...v0.3.1
[v0.3.0]: https://github.com/gnames/gnames/compare/v0.2.2...v0.3.0
[v0.2.2]: https://github.com/gnames/gnames/compare/v0.2.1...v0.2.2
[v0.2.1]: https://github.com/gnames/gnames/compare/v0.2.0...v0.2.1
[v0.2.0]: https://github.com/gnames/gnames/compare/v0.1.8...v0.2.0
[v0.1.8]: https://github.com/gnames/gnames/compare/v0.1.7...v0.1.8
[v0.1.7]: https://github.com/gnames/gnames/compare/v0.1.6...v0.1.7
[v0.1.6]: https://github.com/gnames/gnames/compare/v0.1.5...v0.1.6
[v0.1.5]: https://github.com/gnames/gnames/compare/v0.1.4...v0.1.5
[v0.1.4]: https://github.com/gnames/gnames/compare/v0.1.3...v0.1.4
[v0.1.3]: https://github.com/gnames/gnames/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/gnames/gnames/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/gnames/gnames/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/gnames/gnames/compare/v0.0.4...v0.1.0
[v0.0.4]: https://github.com/gnames/gnames/compare/v0.0.3...v0.0.4
[v0.0.3]: https://github.com/gnames/gnames/compare/v0.0.2...v0.0.3
[v0.0.2]: https://github.com/gnames/gnames/compare/v0.0.1...v0.0.2
[v0.0.1]: https://github.com/gnames/gnames/tree/v0.0.0...v0.0.1
[v0.0.0]: https://github.com/gnames/gnames/tree/v0.0.0

[#100]: https://github.com/gnames/gnames/issues/100
[#99]: https://github.com/gnames/gnames/issues/99
[#98]: https://github.com/gnames/gnames/issues/98
[#97]: https://github.com/gnames/gnames/issues/97
[#96]: https://github.com/gnames/gnames/issues/96
[#95]: https://github.com/gnames/gnames/issues/95
[#94]: https://github.com/gnames/gnames/issues/94
[#93]: https://github.com/gnames/gnames/issues/93
[#92]: https://github.com/gnames/gnames/issues/92
[#91]: https://github.com/gnames/gnames/issues/91
[#90]: https://github.com/gnames/gnames/issues/90
[#89]: https://github.com/gnames/gnames/issues/89
[#88]: https://github.com/gnames/gnames/issues/88
[#87]: https://github.com/gnames/gnames/issues/87
[#86]: https://github.com/gnames/gnames/issues/86
[#85]: https://github.com/gnames/gnames/issues/85
[#84]: https://github.com/gnames/gnames/issues/84
[#83]: https://github.com/gnames/gnames/issues/83
[#82]: https://github.com/gnames/gnames/issues/82
[#81]: https://github.com/gnames/gnames/issues/81
[#80]: https://github.com/gnames/gnames/issues/80
[#79]: https://github.com/gnames/gnames/issues/79
[#78]: https://github.com/gnames/gnames/issues/78
[#77]: https://github.com/gnames/gnames/issues/77
[#76]: https://github.com/gnames/gnames/issues/76
[#75]: https://github.com/gnames/gnames/issues/75
[#74]: https://github.com/gnames/gnames/issues/74
[#73]: https://github.com/gnames/gnames/issues/73
[#72]: https://github.com/gnames/gnames/issues/72
[#71]: https://github.com/gnames/gnames/issues/71
[#70]: https://github.com/gnames/gnames/issues/70
[#69]: https://github.com/gnames/gnames/issues/69
[#68]: https://github.com/gnames/gnames/issues/68
[#67]: https://github.com/gnames/gnames/issues/67
[#66]: https://github.com/gnames/gnames/issues/66
[#65]: https://github.com/gnames/gnames/issues/65
[#64]: https://github.com/gnames/gnames/issues/64
[#63]: https://github.com/gnames/gnames/issues/63
[#62]: https://github.com/gnames/gnames/issues/62
[#61]: https://github.com/gnames/gnames/issues/61
[#60]: https://github.com/gnames/gnames/issues/60
[#59]: https://github.com/gnames/gnames/issues/59
[#58]: https://github.com/gnames/gnames/issues/58
[#57]: https://github.com/gnames/gnames/issues/57
[#56]: https://github.com/gnames/gnames/issues/56
[#55]: https://github.com/gnames/gnames/issues/55
[#54]: https://github.com/gnames/gnames/issues/54
[#53]: https://github.com/gnames/gnames/issues/53
[#52]: https://github.com/gnames/gnames/issues/52
[#51]: https://github.com/gnames/gnames/issues/51
[#50]: https://github.com/gnames/gnames/issues/50
[#49]: https://github.com/gnames/gnames/issues/49
[#48]: https://github.com/gnames/gnames/issues/48
[#47]: https://github.com/gnames/gnames/issues/47
[#46]: https://github.com/gnames/gnames/issues/46
[#45]: https://github.com/gnames/gnames/issues/45
[#44]: https://github.com/gnames/gnames/issues/44
[#43]: https://github.com/gnames/gnames/issues/43
[#42]: https://github.com/gnames/gnames/issues/42
[#41]: https://github.com/gnames/gnames/issues/41
[#40]: https://github.com/gnames/gnames/issues/40
[#39]: https://github.com/gnames/gnames/issues/39
[#38]: https://github.com/gnames/gnames/issues/38
[#37]: https://github.com/gnames/gnames/issues/37
[#36]: https://github.com/gnames/gnames/issues/36
[#35]: https://github.com/gnames/gnames/issues/35
[#34]: https://github.com/gnames/gnames/issues/34
[#33]: https://github.com/gnames/gnames/issues/33
[#32]: https://github.com/gnames/gnames/issues/32
[#31]: https://github.com/gnames/gnames/issues/31
[#30]: https://github.com/gnames/gnames/issues/30
[#29]: https://github.com/gnames/gnames/issues/29
[#28]: https://github.com/gnames/gnames/issues/28
[#27]: https://github.com/gnames/gnames/issues/27
[#26]: https://github.com/gnames/gnames/issues/26
[#25]: https://github.com/gnames/gnames/issues/25
[#24]: https://github.com/gnames/gnames/issues/24
[#23]: https://github.com/gnames/gnames/issues/23
[#22]: https://github.com/gnames/gnames/issues/22
[#21]: https://github.com/gnames/gnames/issues/21
[#20]: https://github.com/gnames/gnames/issues/20
[#19]: https://github.com/gnames/gnames/issues/19
[#18]: https://github.com/gnames/gnames/issues/18
[#17]: https://github.com/gnames/gnames/issues/17
[#16]: https://github.com/gnames/gnames/issues/16
[#15]: https://github.com/gnames/gnames/issues/15
[#14]: https://github.com/gnames/gnames/issues/14
[#13]: https://github.com/gnames/gnames/issues/13
[#12]: https://github.com/gnames/gnames/issues/12
[#11]: https://github.com/gnames/gnames/issues/11
[#10]: https://github.com/gnames/gnames/issues/10
[#9]: https://github.com/gnames/gnames/issues/9
[#8]: https://github.com/gnames/gnames/issues/8
[#7]: https://github.com/gnames/gnames/issues/7
[#6]: https://github.com/gnames/gnames/issues/6
[#5]: https://github.com/gnames/gnames/issues/5
[#4]: https://github.com/gnames/gnames/issues/4
[#3]: https://github.com/gnames/gnames/issues/3
[#2]: https://github.com/gnames/gnames/issues/2
[#1]: https://github.com/gnames/gnames/issues/1

[changelog guidelines]: https://github.com/olivierlacan/keep-a-changelog
[Semantic Versioning guidelines]: https://semver.org/
