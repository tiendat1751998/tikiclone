# Elasticsearch Product Search Index Configurations

To allow lightning-fast search in Vietnamese (tone-marks extraction, autocorrection), the `tiki_products` index must use specialized analysis mappings.

## Elasticsearch Settings & Analysis Mapping (`index_settings.json`)
```json
{
  "settings": {
    "analysis": {
      "filter": {
        "vietnamese_stop": {
          "type": "stop",
          "stopwords": ["và", "hoặc", "cho", "của", "tại", "ở"]
        }
      },
      "analyzer": {
        "vi_analyzer": {
          "type": "custom",
          "tokenizer": "icu_tokenizer",
          "filter": [
            "lowercase",
            "vietnamese_stop",
            "icu_folding"
          ]
        }
      }
    }
  },
  "mappings": {
    "properties": {
      "product_id": { "type": "keyword" },
      "title": {
        "type": "text",
        "analyzer": "vi_analyzer",
        "search_analyzer": "vi_analyzer"
      },
      "description": {
        "type": "text",
        "analyzer": "vi_analyzer"
      },
      "categories": { "type": "keyword" },
      "price": { "type": "double" },
      "sold_count": { "type": "integer" }
    }
  }
}
```
