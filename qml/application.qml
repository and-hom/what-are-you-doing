import QtQuick 2.0
import QtQuick.Controls 2.2
import QtQuick.Controls.Material 2.2
import QtQuick.Controls.Styles 1.4
import QtQuick.Layouts 1.3
import QtQuick.Window 2.0
import Qt.labs.platform 1.0

ApplicationWindow {
    id:appWindow
    visible: true
    title: "What are you doung now?"

    minimumWidth: 300
    maximumWidth: 300

    minimumHeight: Math.min(scrollView.height, 600)
    maximumHeight: 800

    x: Screen.width / 2 - width / 2
    y: Screen.height / 2 - height / 2

    Material.theme: Material.Dark
    Material.accent: Material.Green

    flags: Qt.Window | Qt.WindowCloseButtonHint | Qt.WindowStaysOnTopHint

    onClosing: {
        visible = false
        bridge.windowClosed()
    }

    ScrollView {
        id: scrollView
        width: mainLayout.width
        height: mainLayout.height
        ColumnLayout {
            id: mainLayout
            anchors.fill: parent
            spacing: 4
            ColumnLayout {
                id: selectLayout
                spacing: 4
                Layout.alignment: Qt.AlignTop
                anchors.fill: parent

                ButtonGroup {
                    id: radioGroup
                    onClicked: submit.enabled = true
                }

                Repeater {
                    id: selectorList
                    model: projectList

                    RadioButton {
                        text: display
                        ButtonGroup.group: radioGroup
                    }
                }

                RadioButton {
                    text: "Другое"
                    ButtonGroup.group: radioGroup
                }
            }
            Button {
                id: submit
                text: "Ok!"
                enabled: false
                // todo: set width 100% more smart way
                implicitWidth: 290

                Layout.fillHeight : false
                Layout.fillWidth: true
                Layout.margins: 5

                onClicked:{
                    bridge.okPressed(radioGroup.checkedButton.text)
                    bridge.windowClosed()
                    appWindow.visible = false
                }
            }
        }
    }

    Timer {
        id: timer
        interval: bridge.showPeriod()
        running: true
        repeat: true

        onTriggered: {
            appWindow.visible = true
        }
    }
    Timer {
        id: trayicon_timer
        interval: 600000
        running: true
        repeat: true
        triggeredOnStart: true

        onTriggered: {
            if (bridge.isRedMode()) {
                trayicon.iconSource = "file:images/icon_red.png"
            } else if (bridge.isYellowMode()) {
                trayicon.iconSource = "file:images/icon_yellow.png"
            } else {
                trayicon.iconSource = "file:images/icon_green.png"
            }
        }
    }

    SystemTrayIcon {
        id: trayicon
        visible: true
        iconSource: "file:images/icon.png"

        menu: Menu {
            MenuItem {
                text: qsTr("Copy this week report to clipboard")
                onTriggered: bridge.copyThisWeekPressed()
            }
            MenuItem {
                text: qsTr("Copy prev week report to clipboard")
                onTriggered: bridge.copyPrevWeekPressed()
            }
            MenuItem {
                text: qsTr("Quit")
                onTriggered: Qt.quit()
            }
        }
    }
}
