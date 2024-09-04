# Docs For Foyle

## Hugo Version

* The site currently seems to build with hugo v0.117.0
* With v0.131 we get errors with the template
* See https://github.com/jlewi/foyle/issues/227

## To Add A Banner At the Top Of A Page use

```docsy {"id":"01J49N8FNVE9SFP3DPTM3F52A1"}
{{% pageinfo %}}
This is a placeholder page that shows you how to use this template site.
{{% /pageinfo %}}
```

```sh {"id":"01J49N9KR89D85Z84YDEVAC563"}
To run hugo locally
```

```sh {"id":"01J49NA03YRB6QCR5CGWN31V0K"}
hugo serve -D
```

## References

[stateful/runme#663](https://github.com/stateful/runme/issues/663) - RunMe notebooks don't pell well with Hugo