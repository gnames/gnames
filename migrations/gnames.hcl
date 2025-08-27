table "canonical_fulls" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "name" {
    null = false
    type = character_varying(255)
  }
  primary_key {
    columns = [column.id]
  }
}
table "canonical_stems" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "name" {
    null = false
    type = character_varying(255)
  }
  primary_key {
    columns = [column.id]
  }
}
table "canonicals" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "name" {
    null = false
    type = character_varying(255)
  }
  primary_key {
    columns = [column.id]
  }
}
table "data_sources" {
  schema = schema.public
  column "id" {
    null = false
    type = smallint
  }
  column "uuid" {
    null    = true
    type    = uuid
    default = "00000000-0000-0000-0000-000000000000"
  }
  column "title" {
    null = true
    type = character_varying(255)
  }
  column "title_short" {
    null = true
    type = character_varying(50)
  }
  column "version" {
    null = true
    type = character_varying(50)
  }
  column "revision_date" {
    null = true
    type = text
  }
  column "doi" {
    null = true
    type = character_varying(50)
  }
  column "citation" {
    null = true
    type = text
  }
  column "authors" {
    null = true
    type = text
  }
  column "description" {
    null = true
    type = text
  }
  column "website_url" {
    null = true
    type = character_varying(255)
  }
  column "data_url" {
    null = true
    type = character_varying(255)
  }
  column "outlink_url" {
    null = true
    type = text
  }
  column "is_outlink_ready" {
    null = true
    type = boolean
  }
  column "is_curated" {
    null = true
    type = boolean
  }
  column "is_auto_curated" {
    null = true
    type = boolean
  }
  column "has_taxon_data" {
    null = true
    type = boolean
  }
  column "record_count" {
    null = true
    type = integer
  }
  column "vern_record_count" {
    null = true
    type = integer
  }
  column "updated_at" {
    null = true
    type = timestamp
  }
  primary_key {
    columns = [column.id]
  }
}
table "name_string_indices" {
  schema = schema.public
  column "data_source_id" {
    null = true
    type = integer
  }
  column "record_id" {
    null = true
    type = character_varying(255)
  }
  column "name_string_id" {
    null = true
    type = uuid
  }
  column "outlink_id" {
    null = true
    type = character_varying(255)
  }
  column "global_id" {
    null = true
    type = character_varying(255)
  }
  column "name_id" {
    null = true
    type = character_varying(255)
  }
  column "local_id" {
    null = true
    type = character_varying(255)
  }
  column "code_id" {
    null = true
    type = smallint
  }
  column "rank" {
    null = true
    type = character_varying(255)
  }
  column "taxonomic_status" {
    null = true
    type = character_varying(255)
  }
  column "accepted_record_id" {
    null = true
    type = character_varying(255)
  }
  column "classification" {
    null = true
    type = text
  }
  column "classification_ids" {
    null = true
    type = text
  }
  column "classification_ranks" {
    null = true
    type = text
  }
  index "accepted_record_id" {
    columns = [column.accepted_record_id]
  }
  index "name_string_ids_idx" {
    columns = [column.data_source_id, column.record_id, column.name_string_id]
  }
}
table "name_strings" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "name" {
    null = false
    type = character_varying(500)
  }
  column "year" {
    null = true
    type = integer
  }
  column "cardinality" {
    null = true
    type = integer
  }
  column "canonical_id" {
    null = true
    type = uuid
  }
  column "canonical_full_id" {
    null = true
    type = uuid
  }
  column "canonical_stem_id" {
    null = true
    type = uuid
  }
  column "virus" {
    null = true
    type = boolean
  }
  column "bacteria" {
    null    = false
    type    = boolean
    default = false
  }
  column "surrogate" {
    null = true
    type = boolean
  }
  column "parse_quality" {
    null    = false
    type    = integer
    default = 0
  }
  primary_key {
    columns = [column.id]
  }
  index "canonical" {
    columns = [column.canonical_id]
  }
  index "canonical_full" {
    columns = [column.canonical_full_id]
  }
  index "canonical_stem" {
    columns = [column.canonical_stem_id]
  }
}
table "vernacular_string_indices" {
  schema = schema.public
  column "data_source_id" {
    null = true
    type = integer
  }
  column "record_id" {
    null = true
    type = character_varying(255)
  }
  column "vernacular_string_id" {
    null = true
    type = uuid
  }
  column "language_orig" {
    null = true
    type = character_varying(255)
  }
  column "language" {
    null = true
    type = character_varying(255)
  }
  column "lang_code" {
    null = true
    type = character_varying(3)
  }
  column "locality" {
    null = true
    type = character_varying(255)
  }
  column "country_code" {
    null = true
    type = character_varying(50)
  }
  index "vernacular_string_idx_idx" {
    columns = [column.data_source_id, column.record_id, column.lang_code]
  }
}
table "vernacular_strings" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "name" {
    null = false
    type = character_varying(255)
  }
  primary_key {
    columns = [column.id]
  }
  index "vern_str_name_idx" {
    columns = [column.name]
  }
}
table "word_name_strings" {
  schema = schema.public
  column "word_id" {
    null = false
    type = uuid
  }
  column "name_string_id" {
    null = false
    type = uuid
  }
  column "canonical_id" {
    null = true
    type = uuid
  }
  primary_key {
    columns = [column.word_id, column.name_string_id]
  }
}
table "words" {
  schema = schema.public
  column "id" {
    null = false
    type = uuid
  }
  column "normalized" {
    null = false
    type = character_varying(255)
  }
  column "modified" {
    null = false
    type = character_varying(255)
  }
  column "type_id" {
    null = true
    type = integer
  }
  primary_key {
    columns = [column.id, column.normalized]
  }
  index "words_modified" {
    columns = [column.modified]
  }
}
schema "public" {
  comment = "standard public schema"
}
