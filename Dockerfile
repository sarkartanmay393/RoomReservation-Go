# Using the OS image for Golang enviroment.
FROM golang


RUN mkdir /usr/local/app

# Setting our systems working directory.
WORKDIR /usr/local/app

# Copying all files from present directory to system /app directory as we already set our working directory.
COPY . .

ENV HOST="localhost"
ENV DB_NAME="roomreservation"
ENV DB_USER="postgres"
ENV DB_PASSWORD=""

# Running commands to download module packages and building a exectuable of our project.
RUN go mod download
RUN go build cmd/webapp/*.go
# Soda CLI
RUN go install github.com/gobuffalo/pop/v6/soda@latest

RUN apt-get update
RUN apt-get install postgresql-common -y
RUN sh /usr/share/postgresql-common/pgdg/apt.postgresql.org.sh -y
RUN apt-get install postgresql-14 -y

RUN service postgresql start

RUN su postgres -
RUN createdb roomreservation

RUN soda migrate up

EXPOSE 8080

CMD ["./main"]








