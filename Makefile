LOCAL_DB_DSN_SHARD01:=postgres://postgres:postgres@localhost:5491/postgres
LOCAL_DB_DSN_SHARD02:=postgres://postgres:postgres@localhost:5492/postgres
LOCAL_DB_DSN_SHARD03:=postgres://postgres:postgres@localhost:5493/postgres


.PHONY: db-run
db-run:
	docker run --rm --name shardgo-1-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=postgres -v shardgo_postgres_data_1:/var/lib/postgresql/data -p 5491:5432 -d postgres && \
	docker run --rm --name shardgo-2-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=postgres -v shardgo_postgres_data_2:/var/lib/postgresql/data -p 5492:5432 -d postgres && \
	docker run --rm --name shardgo-3-postgres -e POSTGRES_USER=postgres -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=postgres -v shardgo_postgres_data_3:/var/lib/postgresql/data -p 5493:5432 -d postgres

.PHONY: db-stop
db-stop:
	docker stop shardgo-1-postgres && \
	docker stop shardgo-2-postgres && \
	docker stop shardgo-3-postgres

.PHONY: init-buckets
init-buckets:
	  goose -dir migrations/init_buckets/1 postgres "${LOCAL_DB_DSN_SHARD01}" up && \
	  goose -dir migrations/init_buckets/2 postgres "${LOCAL_DB_DSN_SHARD02}" up && \
	  goose -dir migrations/init_buckets/3 postgres "${LOCAL_DB_DSN_SHARD03}" up

.PHONY: migrate-up
migrate-up:
	  goose -dir migrations/example postgres "${LOCAL_DB_DSN_SHARD01}?search_path=bucket_1" up && \
      goose -dir migrations/example postgres "${LOCAL_DB_DSN_SHARD01}?search_path=bucket_2" up && \
      goose -dir migrations/example postgres "${LOCAL_DB_DSN_SHARD02}?search_path=bucket_3" up && \
      goose -dir migrations/example postgres "${LOCAL_DB_DSN_SHARD02}?search_path=bucket_4" up && \
      goose -dir migrations/example postgres "${LOCAL_DB_DSN_SHARD03}?search_path=bucket_5" up && \
      goose -dir migrations/example postgres "${LOCAL_DB_DSN_SHARD03}?search_path=bucket_6" up

.PHONY: migrate-down
migrate-down:
	goose -dir migrations/example postgres "${LOCAL_DB_DSN_SHARD01}&options=-c%20search_path%3Dbucket_1" down && \
    goose -dir migrations/example postgres "${LOCAL_DB_DSN_SHARD01}&options=-c%20search_path%3Dbucket_2" down && \
    goose -dir migrations/example postgres "${LOCAL_DB_DSN_SHARD02}&options=-c%20search_path%3Dbucket_3" down && \
    goose -dir migrations/example postgres "${LOCAL_DB_DSN_SHARD02}&options=-c%20search_path%3Dbucket_4" down && \
    goose -dir migrations/example postgres "${LOCAL_DB_DSN_SHARD03}&options=-c%20search_path%3Dbucket_5" down && \
    goose -dir migrations/example postgres "${LOCAL_DB_DSN_SHARD03}&options=-c%20search_path%3Dbucket_6" down

.PHONY: lint
lint:
	golangci-lint run \
		--config=.golangci.pipeline.yaml \
		--sort-results \
		--max-issues-per-linter=1000 \
		--max-same-issues=1000 \
		./...
