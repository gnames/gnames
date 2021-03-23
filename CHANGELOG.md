# Changelog

The `gnames` project follows [Semantic Versioning guidelines].

## Unreleased

## [v0.1.8] - 2021-03-22

### Added

- update gnparser to v1.1.0
- stop processing canceled, or expired requests [#65].
- updated dependenies to gnparser v1.0.5, gnmatcher v0.5.5.

## [v0.1.7] - 2021-01-24

### Added

- Update gnparser to v1.0.4 [#62]

## [v0.1.6] - 2020-12-15

### Added

- The score entity is now based on an interface [#59].
- Reduce impact of names with huge number of instances that slowed down
  name verification significantly [#56].
- Improve documentation [#54].

## [v0.1.5] - 2020-12-09

### Added

- Change server timeout for reading and writing to 5 min.

## [v0.1.4] - 2020-12-08

### Added

- Introduce GET method to make ad-hoc verifications easier [#53].
- Score calculation uses parse quality. That helps push names that parsed
  better to the top [#52].

## [v0.1.3] - 2020-12-07

### Added

- New field `isOutlinkReady` for DataSources. This field marks data-sources
  that are prepared to be used as outlinks (for example at BHL) [#47].
- Outlink URLs are now provided in the results [#46].
- DataSource output cleaned up [#45].

## [v0.1.2] - 2020-12-02

### Added

- Set the gnames web service at `https://verifier.globalnames.org`.
- Set dockerhub for releases.

## [v0.1.1] - 2020-12-02

### Added

- Update depenency to `gnlib`.

## [v0.1.0] - 2020-11-22

### Added

- Improve architecture, add OpenAPI [#42].
- Make the code compatible with gnmatcher v0.3.6 [#41].
- Refactor entities, move some of them to gnlib [#40].

## [v0.0.4] - 2020-10-25

### Added

- Increase priority for the authors score [#35].

### Fixed

- 'Acacia vestita may' matches with `PartialExact` to 'Acacia vestita'. Now
  to register Fuzzy there is a limit of 5 characters in a word per edit
  distan event [#38].
- Score calculation uses `edit distance` correctly for fuzzy matches [#36].
- Provide processing of unparseable accepted names [#33].
- Set false positive from gnmatcher as NoMatch. Bloom filters create rare
  false positives. Check every returned name for correctness using Levenshtein
  automata [#32].

## [v0.0.3] - 2020-09-16

### Added

- Get reasonable preferred matches from the real data [#24].
- Reasonable BestMatch from the real data [#28].
- Decrease score for higher edit distance fuzzy matching [#19].
- Currently accepted names generate higher score than synonyms [#20].
- Use curation level for scoring results [#15].
- Use authorship for score calculation [#16].
- Use infrapecific ranks for score calculation [#17].
- Develop a ranking system for score calculation [#18].
- Make gnmatcher functionality interface-based. This allows to choose to
  use gnmatcher as a service or as a library in the future [#23].

## [v0.0.2] - 2020-09-11

### Added

- Improve the code architecture using 'clean architecture' principles [#21].

## [v0.0.1] - 2020-09-05

### Added

- Return complete result [#14].
- Get DataSource metadata out of API [#13].
- Make decode/encode accept either Gob or JSON [#8].
- Migrate from protobuf to Go [#6].
- Incorporate gnmatcher service [#5].
- Send names via HTTP API [#4].
- Setup testing framework [#3].
- Develop a draft ot output format as a protobuffer [#1].

## [v0.0.0] - 2020-05-25

- Add initial commit

## Footnotes

This document follows [changelog guidelines]

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
