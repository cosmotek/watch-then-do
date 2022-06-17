# exec this script using the following command on MacOS `osascript sendMessage.applescript 1234567890 "Hello world"`

on run {targetBuddyPhone, targetMessage}
    tell application "Messages"
        set targetService to 1st service whose service type = iMessage
        set targetBuddy to buddy targetBuddyPhone of targetService
        send targetMessage to targetBuddy
    end tell
end run
