# Sim-Words

## `load` Command

The `load` command is used to ingest a dataset of words, filter and embed them, cluster using k-means, and store the results in the database.

### Usage

```bash
go run . load [flags]
```

### Flags

| Flag              | Shorthand | Default         | Description                                                              |
| ----------------- | --------- | --------------- | ------------------------------------------------------------------------ |
| `-input`         | `-i`      | `""`            | Path to the input file containing the word dataset.                      |
| `-min-index`     | `-mi`     | `100`           | Skip records whose index is less than this value.                        |
| `-min-frequency` | `-mf`     | `10`            | Keep only words with frequency greater than this value.                  |
| `-min-length`    | `-ml`     | `0`             | Keep only words whose length is greater than this value.                 |
| `-k`             | N/A       | `10`            | Number of clusters for k-means.                                          |
| `-kIters`        | N/A       | `1000`          | Maximum number of iterations for k-means.                                |
| `-db`            | N/A       | `"data.sqlite"` | Path to the SQLite database for storing embeddings, clusters, and words. |

### Example

```bash
go run . load \
  -input words.csv \
  -min-index 100 \
  -min-frequency 10 \
  -min-length 2 \
  -k 20 \
  -kIters 500 \
  -db data.sqlite
```

This command will:

1. Load the words from `words.csv`.
2. Filter out words with index < 100, frequency < 10, or length ≤ 2.
3. Compute embeddings for the remaining words.
4. Run k-means clustering with 20 clusters and up to 500 iterations.
5. Store the results in `data.sqlite`.

## `query` Command

The `query` command is used to search for words similar to a given keyword. You can optionally provide a template to embed the keyword in context, and retrieve results based on clusters.

### Usage

```bash
go run . query [flags]
```

### Flags

| Flag   | Shorthand | Default         | Description                                                                                            |
| ------ | --------- | --------------- | ------------------------------------------------------------------------------------------------------ |
| `-k`  | N/A       | `3`             | Select the top `k` clusters most similar to the query.                                                 |
| `-l`  | N/A       | `5`             | From each selected cluster, return the top `l` words most similar to the query.                        |
| `-q`  | N/A       | `""`            | The keyword to query. **This is required.**                                                            |
| `-t`  | N/A       | `""`            | Optional template string for contextualized queries. Use `{{placeholder}}` as the keyword placeholder. |
| `-db` | N/A       | `"data.sqlite"` | Path to the SQLite database containing clusters and word embeddings.                                   |

### Example

#### Simple Query

```bash
go run . query -q apple -k 5 -l 3
```

This command searches for words similar to `"apple"` and retrieves the top 3 words from each of the top 5 clusters.

#### Template Query

```bash
go run . query -q apple -t "I like to eat {{placeholder}}" -k 3 -l 3
```

This command embeds `"apple"` in the template `"I like to eat {{placeholder}}"`, then searches for words in the top 3 clusters and returns the top 3 words per cluster. This allows semantic queries that consider context.

## `serve` Command

The `serve` command starts an HTTP server to handle queries via a REST API. This allows you to run your synonym search service continuously instead of using single commands.

### Usage

```bash
go run . serve [flags]
```

### Flags

| Flag   | Shorthand | Default         | Description                                                          |
| ------ | --------- | --------------- | -------------------------------------------------------------------- |
| `-p`  | N/A       | `3000`          | Port on which the server will listen for HTTP requests.              |
| `-db` | N/A       | `"data.sqlite"` | Path to the SQLite database containing clusters and word embeddings. |

### Example

Start the server on the default port `3000`:

```bash
go run . serve -db data.sqlite
```

Start the server on port `8080`:

```bash
go run . serve -p 8080 -db data.sqlite
```

Once running, you can query the server with HTTP requests:

```http
GET http://localhost:3000/query?q=apple&k=3&l=5
GET http://localhost:3000/query?q=apple&t=I like to eat {{placeholder}}&k=3&l=5
```

*Note: This README was generated with the assistance of AI.*
