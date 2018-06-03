# checkstyle2texterror

Output of checkstyle translator for [reviewdog](https://github.com/haya14busa/reviewdog).

## Usage

    $ cat build/reports/checkstyle/main.xml \
      | checkstyle2texterror \
      | reviewdog -f=golint -name=checkstyle -ci=common

If you want to add 'severity' field on the output:

    $ cat build/reports/checkstyle/main.xml \
      | checkstyle2texterror -s \
      | reviewdog -efm='%f:%l:%c:%t: %m' -name=checkstyle -ci=common

## Options

### Output severity

    -s (-severity) (default: false)

If specified, severity of errors will be outputted.

In that case, you should specify `-efm='%f:%l:%c:%t: %m'` option to reviewdog.
(But I guess severity is not used on reviewdog)

## License

[MIT License](LICENSE.md)
