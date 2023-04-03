.DEFAULT_GOAL := publish
.PHONY: publish build
.SILENT:

publish:
		src=$(word 2,$(MAKECMDGOALS)); \
			fname=$$(basename $$src); \
			echo "src=$$src, fname=$$fname"; \
			docker run -it --rm -v $$src:/srv/$$fname -v /Users/umputun/.ssh/id_rsa:/home/umputun/.ssh/id_rsa uwp-publisher \
				--file=/srv/$$fname --dest=/srv/podcast-uwp/var/media --days=750 --user=umputun; \
			git pull && git commit -am "new podcast $$fname" && git push

build:
		docker build -t uwp-publisher .

local-prep:
		python3 -m venv venv
		. venv/bin/activate && pip install -r requirements.txt
