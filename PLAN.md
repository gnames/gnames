# PLAN: Abbreviated names endpoint

Status: requirements. No implementation decisions are made here.

## Motivation

GNfinder finds many names with an abbreviated genus, such as
`C. elegans` or `P. t. altaica`. One abbreviation alone is hard to
reconcile: `C. alba` matches hundreds of genera. But a text usually
has several abbreviations with the same genus abbreviation, and they
usually belong to the same genus. If gnames receives such names
together and returns candidate genera that explain most of them, the
group gives a strong hint for reconciliation.

The endpoint uses the `abbr_lookups` table built by `gndb optimize`
(see gndb `PLAN.md`). Its keys are the lowercase first letter of the
genus and the epithet stem (`c-elegans`), built by gnparser's
`stemmer.AbbrKey`. Every epithet of a name has its own row, so
`P. t. altaica` finds both *Panthera tigris altaica* and
*Panthera altaica*.

## Context and scope

- The main client is GNfinder, run over about 60 million pages of the
  Biodiversity Heritage Library (BHL). Speed of a single request is
  important.
- Borders of documents are often unknown in BHL. Pages are known to
  come from one book or journal. Deciding which names belong together
  is a GNfinder task, not a gnames task.
- GNfinder knows the taxonomic context of a book or journal from
  GNverifier statistics, and the positions of full genus names in the
  text. It sends this knowledge as hints.
- gnames does not make the final decision. It returns ranked
  candidates with the evidence behind them, and the client decides.

## Endpoint

- `POST /api/v1/abbr_names`: the main endpoint.
- `GET /api/v1/abbr_names/:names`: optional, for manual tests. Names
  are separated by `|`, as in `verifications`.

## Input

| Parameter       | Level    | Description |
|-----------------|----------|-------------|
| `names`         | request  | Abbreviated names as found in text, each with an optional client ID. |
| `genusHints`    | per name | Full genus names found near the abbreviation, the closest first. |
| `mainTaxon`     | request  | Lowest common ancestor of the source's names, from GNverifier. |
| `dominantClass` | request  | Most common class of the source's names, from GNverifier. |
| `tolerance`     | request  | Percentage (0–100) of names that may be unmatched for a genus. |
| `generaLimit`   | request  | Maximum number of candidate genera in the result. |

### names

- Names are sent verbatim. gnames parses them and builds keys, so keys
  are always built by the same gnparser version.
- All names of a request must have the same genus abbreviation, which
  is capitalized and ends with a period (`C.`, `Cae.`). Otherwise the
  request is an error, and no results are returned. This includes
  full genera (`Caenorhabditis`, `Mus`), abbreviations without a
  period or capital (`C`, `cae.`), and different abbreviations (`C.`
  with `Cae.`). Full genera belong to `genusHints`. Authors often use
  a longer abbreviation to distinguish genera with the same first
  letter, so `C.` and `Cae.` should not be mixed.
- An abbreviation with several letters (`Cae.`) limits candidates to
  genera that start with these letters.
- Abbreviated epithets (`P. t. altaica`) are ignored in v1. Only full
  epithets are matched.
- Client IDs allow mapping results back to positions in text, because
  the same verbatim name can occur many times.

### genusHints

- A list of full genera found near the abbreviation, sorted by
  closeness, the closest first. GNfinder decides what "near" means.
- The order is the only distance information. The first hint is the
  strongest.
- A hint matters only if the database has the hinted genus with the
  abbreviation's epithet. This removes most false hints, which are
  common because many genera can start with the same letter.
- Genera that do not start with the abbreviation are ignored.
- The list length is capped.

### mainTaxon and dominantClass

- `mainTaxon` is the lowest taxon that contains most of the names of a
  source. GNverifier calculates it with a threshold, so a few off-topic
  names (host plants in an insect book) do not move it to a very high
  taxon.
- `dominantClass` is the most common class of the source's names. It
  helps when `mainTaxon` is too broad, for example `Chordata` for a book
  about birds and mammals.
- Abbreviated names belong to the dominant group of a source more often
  than other names: authors abbreviate genera they use repeatedly. A
  candidate outside of the dominant group can still win, but needs
  stronger evidence (confirmed hints or several matched names).
- GNverifier statistics return the dominant class, its percentage and
  the distribution of names across classes next to `Kingdom` and
  `MainTaxon` (gnlib v0.65.0, gnstats v0.2.2).

### tolerance

- Misspellings and OCR errors make some names unmatched. A genus is
  still a candidate if the share of unmatched names is within
  `tolerance`.
- A percentage, not a number of names, because groups vary from a few
  names on a page to hundreds in a book.
- The number of allowed unmatched names is rounded down. With 20%, a
  group of 3 names allows none, a group of 5 allows one.
- Default is about 20. `0` requires all names to match.
- A genus needs a minimal number of matched names (about 2, not
  configurable in v1) before unmatched names are tolerated.
- Unmatched names are not corrected by gnames. With a known genus, the
  client can verify the full name (*Caenorhabditis clegans*), where
  fuzzy matching already exists. Unmatched names can also be new
  species that are not in the database.

### generaLimit

- The maximum number of candidate genera in the result. The genera
  with the highest total score are included.

## Matching

- Matching ignores gender endings: keys use epithet stems, so
  *alba*, *albus* and *album* match each other.
- An epithet matches regardless of its position, because species can
  become subspecies and the other way around.
- Nominotypical repeats are removed before comparison, in the input and
  in the database names: if the last epithet is the same as the one
  before it, it is dropped. *Bubo bubo bubo* becomes *Bubo bubo*, and
  botanical autonyms (*Aus bus var. bus*) become species the same way.
  Only epithets are compared, so the genus of *Bubo bubo* stays.
- The last epithet identifies a name. A matched name has one of three
  match types:

  | Type          | Rule                                              | Example input       | Example match                |
  |---------------|---------------------------------------------------|---------------------|------------------------------|
  | full          | the same epithets                                 | `P. tigris altaica` | *Panthera tigris altaica*    |
  |               |                                                   | `B. bubo`           | *Bubo bubo bubo*             |
  | epithet moved | the same last epithet, other epithets differ      | `P. tigris altaica` | *Panthera altaica*           |
  |               |                                                   | `P. altaica`        | *Panthera tigris altaica*    |
  |               |                                                   | `P. tigris altaica` | *Panthera pardus altaica*    |
  | species level | the same species epithet, last epithets differ    | `P. tigris altaica` | *Panthera tigris*            |
  |               |                                                   | `P. tigris`         | *Panthera tigris altaica*    |
  |               |                                                   | `P. tigris altaica` | *Panthera tigris amurensis*  |

  An epithet moves when it changes rank (a subspecies becomes a
  species or the other way around) or species (a subspecies is moved
  to another species).

  A species level match can be coarser than the input, finer than the
  input, or another infraspecies of the same species.

  Rank markers (`var.`, `subsp.`, `f.`) are not part of keys, so a
  change of an infraspecific rank does not change the match type.

- Any other shared epithet is not a match. For example,
  `P. tigris altaica` does not match *Panthera onca tigris*: there
  `tigris` is a subspecies of another species.
- A species level match is not the name from the text, but it still
  shows the genus. It counts for the genus with a smaller weight.
- Hints never exclude candidates. They only change the ranking.

## Score

- The score of a candidate genus is a weighted sum of features
  calculated over all names of the group: matched names, confirmed
  genus hints, unmatched names, and whether the genus is in
  `dominantClass` and in `mainTaxon`.
- A matched name is worth less in a large genus or with a common
  epithet, because such matches happen by chance more often. A large
  genus can still win, but it needs more matched names.
- Enough matched names win over a confirmed hint of another genus.
  `dominantClass` and `mainTaxon` decide only between close candidates.
- A genus without a classification is not penalized as being outside
  of `dominantClass` and `mainTaxon`: an unknown classification is
  neutral, a different one counts against the genus.
- Weights are hardcoded constants, not request parameters. They change
  when tests on real data show it.
- The output returns each part after weighting, so the parts add up to
  the total.
- Details of features and weights are decided at implementation.

## Output

### Metadata

- Input names and their client IDs.
- Input parameters as applied, including defaults.
- Number of names submitted, matched and unmatched.

### Results

A single list of candidate genera, sorted by total score. If nothing
matched, the list is empty and the input names are only in metadata.

Each candidate genus has:

- the genus name and its classification;
- a score: weighted parts for matched names, genus hints, unmatched
  names, dominant class and main taxon, and their total;
- the genus size in distinct epithet stems;
- matched names, each with:
  - the input name and its client ID;
  - the full canonical form it matched;
  - the match type: full, epithet moved or species level;
  - whether the epithet ending matched exactly or only by stem;
- unmatched names (within `tolerance`).

The same input name can be matched by several genera. Names that are
not in any genus's matched list are unmatched by all of them.

## Data

Measured on the `gntest` database (`abbr_lookups` built with gnparser
v1.15.2): 9.06M rows, 499,313 genera, 8.0M stemmed names (6.72M
binomials, 1.27M names with 3 or more words). 78K genera (16%) have
infraspecific names.

Distribution per genus:

| Per genus         | 50% | 90% | 95% | 99% | 99.9% | 99.99% | Max    |
|-------------------|-----|-----|-----|-----|-------|--------|--------|
| binomials         | 2   | 21  | 46  | 200 | 967   | 2,965  | 12,321 |
| trinomials        | 0   | 2   | 6   | 40  | 325   | 1,577  | 17,604 |
| all names         | 2   | 24  | 53  | 240 | 1,251 | 3,914  | 29,925 |
| distinct epithets | 2   | 23  | 49  | 215 | 1,053 | 3,178  | 17,013 |

Genera within a given size, for a future limit of the misspelling check
(see "Out of scope for v1"):

| Size limit   | Genera within | Genera above |
|--------------|---------------|--------------|
| 100          | 97.54%        | 12,280       |
| 500          | 99.67%        | 1,627        |
| 1,000        | 99.89%        | 547          |
| 2,000        | 99.97%        | 144          |
| 5,000        | ~100%         | 13           |

The largest genera by distinct epithets are *Hieracium* (17,013),
*Helix* (9,639), *Rubus* (8,836), *Carabus* (7,394) and *Astragalus*
(6,758). Some of them (*Helix*, *Phalaena*, *Unio*) were large
catch-all genera of the 18th and 19th centuries, and old literature
uses their names.

### Genus table from gndb

`gndb optimize` builds `abbr_genera` in its abbreviation step, right
after `abbr_lookups`. It has one row per genus and kingdom:

| Column                 | Description |
|------------------------|-------------|
| `genus`                | Genus name, as in `abbr_lookups`. |
| `kingdom`              | Kingdom from the classification, empty if there is no kingdom. |
| `epithets_num`         | Number of distinct epithet stems of the genus. |
| `names_num`            | Number of records of the genus in this kingdom. |
| `classification`       | Classification path, ending with the genus. |
| `classification_ranks` | Ranks of the classification. |
| `data_source_id`       | Source of the classification, 0 if there is none. |

- The table is denormalized on purpose. Counting the size at request
  time would scan all rows of a genus (more than 17,000 for
  *Hieracium*). With about 506K rows it is small enough for gnames to
  load into memory at startup. It takes about 50 seconds to build.
- The size is used by the score. All names count: the database covers
  250 years of nomenclature, and its data does not show which names are
  accepted, synonyms, or forgotten.
- Classifications are needed for the output, and for comparison with
  `mainTaxon` and `dominantClass`.

#### Source of classifications

- Classifications come from Catalogue of Life XR (data source 13). It
  is less accurate than Catalogue of Life (data source 1), but covers
  more genera: 63% of `abbr_lookups` genera against 54%. Without CoL XR
  gndb uses Catalogue of Life, and then the largest data source with
  genus ranks.
- The classification of a genus is cut from the classifications of its
  species and lower names: the part above the genus rank, and the
  genus itself. Synonyms have the classification of their current
  name, so a synonym's own genus replaces the genus of the path. Paths
  of accepted names are preferred, then the most common path.
- 37% of genera have no classification in CoL XR. For them the score
  treats `dominantClass` and `mainTaxon` as unknown, not as outside of
  them.
- `mainTaxon` and `dominantClass` names must come from the same
  classification as the genus table, because sources name higher taxa
  differently (Animalia and Metazoa). Verification statistics use
  Catalogue of Life now. An option to choose their data source is
  planned separately.

#### Homonyms

Homonyms are allowed only between nomenclatural codes, such as *Morus*
(a plant and a bird). The table separates them by kingdom, not by code:

- Codes of records are sometimes wrong: CoL XR has about 6,300 records
  whose code contradicts their classification, such as beetles marked
  as botanical. Grouping by code made false homonyms of *Carabus* and
  *Agaricus*.
- Chromista and Protozoa have no fixed code, so the same foraminiferan
  genus can be filed under both codes.

Grouped by kingdom, 6,173 genera (2%) have more than one row. Many
have a small second side that is still a real homonym, such as
*Moseleya* (a coral and a plant) or *Gomphos* (a fish and a fungus),
so no threshold is used.

v1 handles homonyms simply:

- All rows of a homonym have the same size, counting names of all its
  kingdoms. The smaller homonym looks larger than it is: *Ficus* the
  snail (133 names) gets the size of *Ficus* the plant (3,736). It can
  still win, but needs more matched names.
- `dominantClass` and `mainTaxon` are true if any classification of the
  genus fits. *Ficus* in a book about molluscs gets the bonus. The
  output shows all classifications, so the client sees that the genus
  is ambiguous.

## Out of scope for v1

- Misspellings. Unmatched names are not compared with other epithets
  of a candidate genus. They count against `tolerance`, and the client
  can verify the full name with fuzzy matching once it knows the genus.
  - A comparison needs all epithet stems of a genus, but `abbr_lookups`
    is indexed only by key. Options for later: an index on
    `abbr_lookups (genus) INCLUDE (canonical_stem)` and one more query
    per request, or epithet stems of genera loaded into memory (about
    8M stems, a few hundred MB). The choice depends on speed and memory
    measured on BHL data.
  - Large genera give many false matches for misspellings, so the check
    would come with a request parameter `maxGenusSize`: the largest
    genus, in distinct epithet stems, that is checked. Larger genera get
    exact matches only, and the result marks them. A default of 500
    covers 99.67% of genera (see "Data").
- Separate sizes for homonyms of different kingdoms. It needs a kingdom
  in every `abbr_lookups` row (see "Future" in gndb `PLAN.md`).
- Checks of gnparser versions of gnames and `abbr_lookups`. Keys depend
  only on `stemmer.AbbrKey`, which has not changed. If it changes,
  gnames can compare the versions once at startup and log a warning.
- Abbreviated epithets (`P. t. altaica`) used as a filter.
- Several abbreviation groups in one request, if throughput requires
  it.
- Correction of OCR errors in the genus abbreviation (`C.` read as
  `G.`).
- Ranking by taxonomic data stored in `abbr_lookups` (see "Future" in
  gndb `PLAN.md`).

## Open questions

- Weights of score parts. They should be tuned on a test set of
  abbreviated names from BHL that are resolved manually.
- Maximum number of names per request and of genus hints per name.
