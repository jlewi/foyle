Fetch the hydros logs for the build of the image

To fetch the hydros logs for the build of the image `vscode-ext`, you can follow these steps:

1. Use `gcloud` to fetch the hydros logs for the image `vscode-ext`:

```sh {"id":"01HZ3BDKKQK01SE8K772C2D23B"}
gcloud logging read 'logName="projects/foyle-dev/logs/hydros" jsonPayload.image="vscode"' --freshness=14d --project=foyle-dev
```

2. Check the RunMe logs to ensure the command was logged:

```text {"id":"01HZ3BDKKQK01SE8K777TC66VQ"}
ls -ltr "/Users/jlewi/Library/Application Support/runme/logs/"
```

3. If the logs weren't recorded, verify if RunMe is running with the logs flag:

```text {"id":"01HZ3BDKKQK01SE8K77DSWB39M"}
ps -ef | grep runme
```

4. If RunMe is not running with the correct configuration, update the RunMe extension to enable logs:

- Bump the package extension version and reinstall it.
- Confirm that RunMe is now running with `ai-logs=true`.

5. Retry fetching the hydros logs for the `vscode-ext` image:

```text {"id":"01HZ3BDKKQK01SE8K77MZDZX7A"}
gcloud logging read 'logName="projects/foyle-dev/logs/hydros" jsonPayload.image="vscode-ext"' --freshness=14d --project=foyle-dev
```

6. Check the logs directory again for any recent logs:

```text {"id":"01HZ3BDKKQK01SE8K77QG68VVX"}
ls -ltr "/Users/jlewi/Library/Application Support/runme/logs/"
```

7. Look for the most recent log associated with a specific log ID (e.g., `01HYZXS2Q5XYX7P3PT1KH5Q881`):

```text {"id":"01HZ3BDKKQK01SE8K77XFD6D4G"}
grep -r 01HYZXS2Q5XYX7P3PT1KH5Q881 "/Users/jlewi/Library/Application Support/runme/logs/"
```

Based on the output you provided, it appears that you have already performed some of these steps. If you encounter any issues or need further assistance, feel free to ask!