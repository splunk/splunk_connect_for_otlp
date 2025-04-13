# Releasing

This repository releases the addon and a git tag manually.

Steps:
1. Update the version in ta/otlpinput/default/app.conf
2. Commit the change
3. Create a release tag `vx.x.x`
4. Push the tag
5. Make the addon `make tgz`
6. Create the github release, add otlpinput.tgz file to the release artifacts.
