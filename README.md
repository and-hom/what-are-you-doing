## How to build

1. Download and unpack QT 5.9
```
wget http://download.qt.io/official_releases/qt/5.9/5.9.1/qt-opensource-linux-x64-5.9.1.run
chmod +x qt-opensource-linux-x64-5.9.1.run
./qt-opensource-linux-x64-5.9.1.run
```
2. Set the environment variables:
```
export QT_VERSION=5.9.1
export QT_DIR=$HOME/Qt5.9.1/
export QT_QMAKE_DIR=$QT_DIR/5.9.1/gcc_64/bin
```
3. Get https://github.com/therecipe/qt:
```
go get github.com/therecipe/qt
```
4. Install therecipe/qt tools:
```
go install github.com/therecipe/qt/cmd/qtmoc
go install github.com/therecipe/qt/cmd/qtrcc
go install github.com/therecipe/qt/cmd/qtsetup
```
5. Build QT bindings
```
qtsetup
```
6. Build and install app
```
go get github.com/and-hom/what-are-you-doing
cd $GOPATH/src/github.com/and-hom/what-are-you-doing

qtrcc
qtmoc

# may take several minutes or more!
go build

go install
```
7. Install https://glide.sh/
8. Install https://github.com/mh-cbon/go-bin-deb
9. Build a package
```
rm -rf ./pkg-build
rm -f *.deb
go-bin-deb generate

```

## Configuration
You can see/edit configuration file ``/etc/what-are-you-doing/config.yaml`` and override it with ``%s/.what-are-you-doing/config.yaml``

## Usage
1. ``what-are-you-doing`` - start main loop and show ask window every period (see **Configuration**)
2. ``what-are-you-doing print`` - print this week report (if tray menu is unavailable)
2. ``what-are-you-doing print --normalized`` - print this week report, but use 40 hours as 100%
3. ``what-are-you-doing print --prev`` - print previous week report (if tray menu is unavailable)
3. ``what-are-you-doing print --prev`` - print previous week report (if tray menu is unavailable), but use 40 hours as 100%
3. ``what-are-you-doing --help`` - see options