# Security Policy

## Supported versions

Security fixes are released for the latest minor version of the latest major version. Older versions do not receive fixes; upgrade to the latest release.

## Reporting a vulnerability

Please do not open a public issue for a security problem.

Report it privately through GitHub: open the repository's **Security** tab and choose **Report a vulnerability** ([direct link](https://github.com/terrakube-io/terrakube-go/security/advisories/new)). Include the library version, what an attacker can do, and steps or code that show the problem.

You can expect a first answer within a week. Once a fix is released, the advisory is published and you are credited unless you ask not to be.

## Scope

This policy covers the Go client in this repository: how it builds requests, handles the bearer token, and decodes responses. A vulnerability in the Terrakube server belongs to the [Terrakube project](https://github.com/terrakube-io/terrakube).
