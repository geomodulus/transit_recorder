# transit_recorder

Record TTC route vehicle locations for replay.

## Usage

To record vehicle locations for a route:

```
%> go run . --route 510,504 record
Recording route "510"...
Recording route "504"...
510 updated with 11 active vehicles
504 updated with 26 active vehicles
...
```

To export them later:

```
%> go run . --routes 504,506 --start 2023-05-01T10:00:00Z --end 2023-05-01T14:00:00Z export
```

That will export the recorded locations into a JSON file.
