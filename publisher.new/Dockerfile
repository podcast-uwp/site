FROM python:3.11-alpine

ENV REMOTE_HOST=podcast.umputun.com

# Install required packages
RUN apk add --update --no-cache ansible openssh-client && \
    rm -rf /var/cache/apk/*

# Copy files into the container
COPY publisher.py /srv/publisher.py
COPY ansible.yml /srv/ansible.yml
COPY requirements.txt /srv/requirements.txt
COPY cover.jpg /srv/cover.jpg

# Create a virtual environment and install dependencies
RUN pip install --no-cache-dir -r /srv/requirements.txt

# Set the working directory
WORKDIR /srv

# Add user umputun
RUN adduser -D umputun

# Set the owner of the files to the umputun user
RUN chown -R umputun:umputun /srv

# Add SSH known_hosts
RUN mkdir -p /home/umputun/.ssh && ssh-keyscan $REMOTE_HOST >> /home/umputun/.ssh/known_hosts && \
    echo "StrictHostKeyChecking no" >> /home/umputun/.ssh/config

# Set the entrypoint to run the process under the umputun user
USER umputun
ENTRYPOINT ["python", "publisher.py", "run-all", "--host=$REMOTE_HOST"]
