FROM byuoitav/amd64-alpine
MAINTAINER Daniel Randall <danny_randall@byu.edu>

ARG NAME
ENV name=${NAME}

COPY ${name}-bin ${name}-bin 
COPY version.txt version.txt
COPY config.json config.json

# add any required files/folders here
COPY dash-dist dash-dist

ENTRYPOINT ./${name}-bin
