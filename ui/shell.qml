import Quickshell
import Quickshell.Io
import QtQuick
import QtQuick.Layouts
import QtQuick.Controls
import QtQuick.Shapes

ShellRoot {
    PanelWindow {
        id: panel

        anchors.right: true
        anchors.top: true
        anchors.bottom: true
        margins.right: 16
        margins.top: 24
        margins.bottom: 24

        implicitWidth: 400

        color: "transparent"
        surfaceFormat.opaque: false
        focusable: true
        aboveWindows: true
        exclusiveZone: 0

        property var accounts: []
        property string query: ""
        property int tick: Math.floor(Date.now() / 1000)

        property var filtered: {
            const q = query.toLowerCase();
            if (q === "") return accounts;
            return accounts.filter(a =>
                a.issuer.toLowerCase().includes(q) ||
                a.label.toLowerCase().includes(q)
            );
        }

        function avatarColor(name) {
            const palette = ["#6366f1", "#8b5cf6", "#ec4899", "#14b8a6", "#f59e0b", "#22c55e", "#3b82f6", "#ef4444"];
            let h = 0;
            for (let i = 0; i < name.length; i++)
                h = name.charCodeAt(i) + ((h << 5) - h);
            return palette[Math.abs(h) % palette.length];
        }

        Timer {
            interval: 1000
            running: true
            repeat: true
            onTriggered: panel.tick = Math.floor(Date.now() / 1000)
        }

        Timer {
            interval: 30000
            running: true
            repeat: true
            onTriggered: listProc.running = true
        }

        Process {
            id: listProc
            command: ["totp", "list"]
            property string buf: ""
            stdout: SplitParser {
                onRead: line => listProc.buf += line
            }
            onExited: (code) => {
                if (code === 0 && listProc.buf !== "") {
                    try { panel.accounts = JSON.parse(listProc.buf); } catch (e) {}
                }
                listProc.buf = "";
            }
        }

        Process {
            id: copyProc
            onExited: (code) => { if (code === 0) closeTimer.start(); }
        }

        Timer {
            id: closeTimer
            interval: 180
            repeat: false
            onTriggered: Qt.quit()
        }

        function copyFocused() {
            const idx = accountList.currentIndex;
            if (idx < 0 || idx >= panel.filtered.length) return;
            copyProc.command = ["totp", "copy", panel.filtered[idx].id];
            copyProc.running = true;
        }

        Component.onCompleted: listProc.running = true

        Item {
            anchors.fill: parent
            clip: true

            Rectangle {
                id: chrome

                property bool opened: false

                width: parent.width
                height: parent.height
                x: opened ? 0 : parent.width + 40
                radius: 18
                border.width: 1
                border.color: "#27272a"

                gradient: Gradient {
                    orientation: Gradient.Vertical
                    GradientStop { position: 0.0; color: "#1e1e2c" }
                    GradientStop { position: 0.5; color: "#18181f" }
                    GradientStop { position: 1.0; color: "#141418" }
                }

                Behavior on x {
                    NumberAnimation { duration: 220; easing.type: Easing.OutCubic }
                }

                Component.onCompleted: {
                    opened = true;
                    searchField.forceActiveFocus();
                }

                ColumnLayout {
                    anchors.fill: parent
                    anchors.margins: 20
                    spacing: 0

                    // Header
                    RowLayout {
                        Layout.fillWidth: true
                        Layout.bottomMargin: 16

                        Column {
                            spacing: 3
                            Text {
                                text: "Authenticator"
                                color: "#fafafa"
                                font.pixelSize: 20
                                font.weight: Font.Bold
                            }
                            Text {
                                text: panel.filtered.length + " of " + panel.accounts.length + " accounts"
                                color: "#52525b"
                                font.pixelSize: 12
                            }
                        }

                        Item { Layout.fillWidth: true }

                        Rectangle {
                            width: 8; height: 8; radius: 4
                            color: "#6366f1"
                        }
                    }

                    // Divider
                    Rectangle {
                        Layout.fillWidth: true
                        height: 1
                        color: "#27272a"
                        Layout.bottomMargin: 16
                    }

                    // Search
                    Rectangle {
                        Layout.fillWidth: true
                        height: 44
                        radius: 12
                        color: "#111117"
                        border.width: 1
                        border.color: searchField.activeFocus ? "#6366f1" : "#27272a"
                        Layout.bottomMargin: 16

                        Behavior on border.color { ColorAnimation { duration: 150 } }

                        RowLayout {
                            anchors { fill: parent; leftMargin: 14; rightMargin: 14 }
                            spacing: 10

                            Text {
                                text: "⌕"
                                color: searchField.activeFocus ? "#6366f1" : "#3f3f46"
                                font.pixelSize: 20
                                Behavior on color { ColorAnimation { duration: 150 } }
                            }

                            Item {
                                Layout.fillWidth: true
                                Layout.fillHeight: true

                                TextInput {
                                    id: searchField
                                    anchors { fill: parent }
                                    color: "#fafafa"
                                    font.pixelSize: 14
                                    verticalAlignment: TextInput.AlignVCenter
                                    focus: true
                                    onTextChanged: panel.query = text

                                    Keys.onPressed: (event) => {
                                        switch (event.key) {
                                        case Qt.Key_Escape:
                                            Qt.quit();
                                            event.accepted = true;
                                            break;
                                        case Qt.Key_Return:
                                        case Qt.Key_Enter:
                                            panel.copyFocused();
                                            event.accepted = true;
                                            break;
                                        case Qt.Key_Up:
                                            accountList.decrementCurrentIndex();
                                            event.accepted = true;
                                            break;
                                        case Qt.Key_Down:
                                            accountList.incrementCurrentIndex();
                                            event.accepted = true;
                                            break;
                                        case Qt.Key_Space:
                                            if (searchField.text === "") {
                                                panel.copyFocused();
                                                event.accepted = true;
                                            }
                                            break;
                                        case Qt.Key_J:
                                            if (event.modifiers === Qt.NoModifier) {
                                                accountList.incrementCurrentIndex();
                                                event.accepted = true;
                                            }
                                            break;
                                        case Qt.Key_K:
                                            if (event.modifiers === Qt.NoModifier) {
                                                accountList.decrementCurrentIndex();
                                                event.accepted = true;
                                            }
                                            break;
                                        }
                                    }
                                }

                                Text {
                                    visible: searchField.text === ""
                                    anchors.verticalCenter: parent.verticalCenter
                                    text: "Search accounts…"
                                    color: "#3f3f46"
                                    font.pixelSize: 14
                                }
                            }
                        }
                    }

                    // Account list + empty state
                    Item {
                        Layout.fillWidth: true
                        Layout.fillHeight: true

                        Text {
                            anchors.centerIn: parent
                            visible: panel.filtered.length === 0
                            text: panel.accounts.length === 0
                                ? "No accounts imported"
                                : "No results for \"" + panel.query + "\""
                            color: "#3f3f46"
                            font.pixelSize: 14
                        }

                        ListView {
                            id: accountList
                            anchors.fill: parent
                            clip: true
                            spacing: 8
                            currentIndex: 0
                            model: panel.filtered
                            visible: panel.filtered.length > 0

                            delegate: Rectangle {
                                id: card

                                required property var modelData
                                required property int index

                                width: accountList.width
                                height: 68
                                radius: 12
                                color: "transparent"

                                Rectangle {
                                    anchors.fill: parent
                                    radius: parent.radius
                                    color: accountList.currentIndex === index ? "#1e1e35" : "#111117"
                                    border.width: 1
                                    border.color: accountList.currentIndex === index ? "#6366f1" : "#27272a"
                                    Behavior on color { ColorAnimation { duration: 80 } }
                                    Behavior on border.color { ColorAnimation { duration: 80 } }
                                }

                                Rectangle {
                                    anchors.fill: parent
                                    radius: parent.radius
                                    visible: hov.containsMouse && accountList.currentIndex !== index
                                    color: "#1a1a22"
                                }

                                property real remaining: modelData.period - (panel.tick % modelData.period)
                                property real fraction: remaining / modelData.period
                                property color arcColor: fraction > 0.5 ? "#22c55e" : fraction > 0.25 ? "#f59e0b" : "#ef4444"

                                RowLayout {
                                    anchors { fill: parent; margins: 14 }
                                    spacing: 12

                                    // Avatar
                                    Rectangle {
                                        width: 38
                                        height: 38
                                        radius: 19
                                        color: panel.avatarColor(card.modelData.issuer || card.modelData.label)

                                        Text {
                                            anchors.centerIn: parent
                                            text: (card.modelData.issuer || card.modelData.label).charAt(0).toUpperCase()
                                            color: "#ffffff"
                                            font.pixelSize: 16
                                            font.weight: Font.DemiBold
                                        }
                                    }

                                    // Names
                                    Column {
                                        Layout.fillWidth: true
                                        spacing: 4

                                        Text {
                                            text: card.modelData.issuer || card.modelData.label
                                            color: "#fafafa"
                                            font.pixelSize: 14
                                            font.weight: Font.DemiBold
                                            elide: Text.ElideRight
                                            width: parent.width
                                        }

                                        Text {
                                            text: card.modelData.label
                                            color: "#71717a"
                                            font.pixelSize: 12
                                            elide: Text.ElideRight
                                            width: parent.width
                                            visible: card.modelData.issuer !== "" &&
                                                     card.modelData.issuer !== card.modelData.label
                                        }
                                    }

                                    // Progress arc
                                    Item {
                                        width: 40
                                        height: 40

                                        Shape {
                                            anchors.fill: parent
                                            ShapePath {
                                                strokeWidth: 3
                                                strokeColor: "#27272a"
                                                fillColor: "transparent"
                                                PathAngleArc {
                                                    centerX: 20; centerY: 20
                                                    radiusX: 16; radiusY: 16
                                                    startAngle: -90
                                                    sweepAngle: 360
                                                }
                                            }
                                        }

                                        Shape {
                                            anchors.fill: parent
                                            ShapePath {
                                                strokeWidth: 3
                                                strokeColor: card.arcColor
                                                fillColor: "transparent"
                                                PathAngleArc {
                                                    centerX: 20; centerY: 20
                                                    radiusX: 16; radiusY: 16
                                                    startAngle: -90
                                                    sweepAngle: card.fraction * 360
                                                }
                                            }
                                        }

                                        Text {
                                            anchors.centerIn: parent
                                            text: Math.ceil(card.remaining)
                                            color: card.arcColor
                                            font.pixelSize: 10
                                            font.weight: Font.DemiBold
                                        }
                                    }
                                }

                                MouseArea {
                                    id: hov
                                    anchors.fill: parent
                                    hoverEnabled: true
                                    cursorShape: Qt.PointingHandCursor
                                    onClicked: {
                                        accountList.currentIndex = card.index;
                                        panel.copyFocused();
                                    }
                                }
                            }
                        }
                    }
                }
            }
        }
    }
}
