
.PHONY: db
db:
	docker run --rm -d \
	    --publish=7474:7474 --publish=7687:7687 \
	    --env NEO4J_AUTH=none \
	    --name neo4j \
	    neo4j:4.1.3

.PHONY: stop
stop:
	docker stop \
	    neo4j
