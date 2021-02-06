# RpiCli
RpiCli is a command line tool for server interactions through gRPC


## Currently supported options
```
./client --help
Usage of ./client:
  -fileDownload string
    	Download file. Example: --fileDownload=file.extension
  -fileInput string
    	Run bash from input file. Example: --fileInput=script.sh
  -fileOutput string
    	Save output to file. Example: --fileOutput=example.txt
  -fileUpload string
    	Upload file. Example: --fileUpload=/path/to/file
  -interactive
    	If set, run client in an interactive mode.
  -savePath string
    	Path for file saving. Example: --savePath=/path/to/destination
```

## How to run Server?
```
docker-compose up -d
```