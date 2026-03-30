# Interface Contract

## fetcher.fetch_records() -> list[dict]
Returns raw records from data/input.json as a list of dicts. No transformation.

## transformer.transform_records(records: list[dict]) -> list[dict]
Accepts raw records. Returns transformed records with:
- `id`: int (unchanged)
- `name`: str (stripped and title-cased)
- `score`: int (parsed from raw_score)
- `active`: bool (unchanged)

## pipeline.run() -> None
Calls fetcher.fetch_records(), passes result to transformer.transform_records(),
writes output to data/output.json as a JSON array.
