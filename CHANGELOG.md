# Changelog

## Unreleased

## [v0.1.4]

- Add [#53]: Add GET for verificaitons for easy API querying.
- Add [#52]: Add parse quality to score calculation.

## [v0.1.3]

- Add [#47]: Add isOutlinkReady field for DataSources.
- Add [#46]: Add Outlink field to results.
- Add [#45]: Tidy up DataSource output.

## [v0.1.2]

- Add web service at `https://verifier.globalnames.org`
- Add dockerhub for releases

## [v0.1.1]

- Add compaibility to most recent gnlib

## [v0.1.0]

- Add [#42]: improve architecture, add OpenAPI.
- Add [#41]: make compatible with gnmatcher v0.3.6.
- Add [#40]: Refactor entities, move some of them to gnlib.

## [v0.0.4]

- Add [#35]: Give higher priority for authors score.
- Add [#34]: Add UUIDv5 library to generate IDs for globalnames.org.
- Fix [#38]: PartialExact instead of Fuzzy matchtype for
             'Acacia vestita may'.
- Fix [#36]: 'edit distance' score for fuzzy matches.
- Fix [#33]: Process unparseable accepted names.
- Fix [#32]: Set false positives from gnmatcher as NoMatch.

## [v0.0.3]

- Add [#24]: Get reasonable preferred matches from real data.
- Add [#28]: Get reasonable best matches from real data.
- Add [#19]: Fuzzy matching with higher edit distance score lower.
- Add [#20]: Currently accepted names score higher than synonyms.
- Add [#15]: Use curation level for scoring results.
- Add [#16]: Use authorship for scoring results.
- Add [#17]: Use infraspecies ranks for scoring results.
- Add [#18]: Develop a ranking system.
- Add [#23]: Make gnmatcher access more flexible.

## [v0.0.2]

- Add [#21]: Clean up architecture.

## [v0.0.1]

- Add [#14]: Return complete result.
- Add [#13]: Get DataSource metadata out of API.
- Add [#8]: Make decode/encode accept either Gob or JSON.
- Add [#6]: Migrate from protobuf to Go.
- Add [#5]: Incorporate gnmatcher service.
- Add [#4]: Send names via HTTP API.
- Add [#3]: Setup testing framework.
- Add [#1]: Develop a draft ot output format as a protobuffer.

## Footnotes

This document follows [changelog guidelines]

[v0.1.4]: https://github.com/gnames/gnfinder/compare/v0.1.3...v0.1.4
[v0.1.3]: https://github.com/gnames/gnfinder/compare/v0.1.2...v0.1.3
[v0.1.2]: https://github.com/gnames/gnfinder/compare/v0.1.1...v0.1.2
[v0.1.1]: https://github.com/gnames/gnfinder/compare/v0.1.0...v0.1.1
[v0.1.0]: https://github.com/gnames/gnfinder/compare/v0.0.4...v0.1.0
[v0.0.4]: https://github.com/gnames/gnfinder/compare/v0.0.3...v0.0.4
[v0.0.3]: https://github.com/gnames/gnfinder/compare/v0.0.2...v0.0.3
[v0.0.2]: https://github.com/gnames/gnfinder/compare/v0.0.1...v0.0.2
[v0.0.1]: https://github.com/gnames/gnames/tree/v0.0.1

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
[gnindex]: https://index.globalnames.org
[Ruby gem gndinder]: https://github.com/GlobalNamesArchitecture/gnfinder
