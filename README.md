# eseed

## Debug in VSCode

Copy ***env/example.env -> env/config.env***

Create folder ***.vscode***, create file **launch.json** in ***.vscode***

```
{
	"version": "0.2.0",
	"configurations": [
		{
			"name": "eseed",
			"type": "go",
			"request": "launch",
			"mode": "debug",
			"program": "${workspaceFolder}/main.go"
		}
	]
}
```

# Compile

```
GOOS=linux GOARCH="amd64" go build -o eseed main.go
GOOS=darwin GOARCH="amd64" go build -o eseed main.go
GOOS=windows GOARCH="amd64" go build -o eseed_64bit.exe main.go
GOOS=windows GOARCH="386" go build -o eseed_32bit.exe main.go
```

# Run

```
MacOS, Linux:
  ./eseed -config=env/config.env

Windows:
  eseed_32bit.exe -config=env/config.env
  eseed_64bit.exe -config=env/config.env

Nohup Command
  Start:	nohup ./eseed -config=env/config.env > eseed.log &
  Stop:		pkill eseed
```

## Deploy:

```
brew install make (macos)

apt install make (ubuntu)

make image
```

## Docker:

```
Build:		docker build -t eseed .

Run:		docker run --name eseed -dp 3000:3000 eseed

Log:		docker logs -f eseed
```

## Docker Compose:

```
Start:		docker-compose up -d

Stop:		docker-compose down

Log:		docker logs -f eseed
```