.DEFAULT_GOAL := help
.PHONY: publish build tag upload git-push prepare
.SILENT:

## publish: set mp3 tags and deploy podcast file to the server and push changes to github
# usage: make publish /foo/bar/podcast/ump_podcast123.mp3
publish: tag upload git-push

## prepare: prepares hugo content for the new podcast
# usage: make prepare
prepare:
		@docker rm -f uwp_publisher_tmp
		docker run -d --rm --name=uwp_publisher_tmp uwp-publisher /srv/publisher sleepy --duration=30s
		docker cp uwp_publisher_tmp:/srv/publisher.osx /tmp/uwp-publisher
		/tmp/uwp-publisher prep --location=/Users/umputun/dev.umputun/podcast-uwp/hugo/content/posts
		@docker rm -f uwp_publisher_tmp

## build: build docker image
build:
		cd build && docker build -t uwp-publisher .

## tag: set mp3 tags
tag:
		# upload the file to the server
		src=$(word 2,$(MAKECMDGOALS)); \
  		fname=$$(basename $$src); \
		echo "src=$$src, fname=$$fname"; \
		docker run -it --rm -e APP_UID=502 -v $$src:/srv/$$fname -v /Users/umputun/.ssh/id_rsa:/home/app/.ssh/id_rsa uwp-publisher \
			/srv/publisher mp3 --file=/srv/$$fname

## upload: upload podcast file to the primary server and to the archive server
upload:
		# deploy the file to the podcast server
		src=$(word 2,$(MAKECMDGOALS)); \
		fname=$$(basename $$src); \
		echo "src=$$src, fname=$$fname"; \
		docker run -it --rm -e APP_UID=502 -v $$src:/srv/$$fname -v /Users/umputun/.ssh/id_rsa:/home/app/.ssh/id_rsa uwp-publisher \
			/srv/publisher deploy --file=/srv/$$fname --days-keep=750

## git-push: push changes to github if anything changed
git-push:
		# push changes to github if anything changed
		git pull
		@status=$$(git status --porcelain); \
		if test -n "$$status"; then \
			echo "Changes found:"; \
			echo "$$status"; \
			git add .; \
			git commit -m "auto-update $(shell date)"; \
			git push; \
		else \
			echo "No changes found"; \
		fi

help: Makefile
		@sed -n 's/^##//p' $<
