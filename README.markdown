# vsphere-monitor

vsphere-monitor is a tool used to monitor vSphere and report metrics to Librato.

## Building

First, install [gvt](https://github.com/FiloSottile/gvt) if you don't have it already:

    $ go get -u github.com/FiloSottile/gvt

Download dependencies:

    $ gvt restore

Build:

    $ go install .

## License

vsphere-monitor is released under the MIT license, see the `LICENSE` file for more information.
