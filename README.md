How to build

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
qtrcc
qtmoc
go build
go install
```
