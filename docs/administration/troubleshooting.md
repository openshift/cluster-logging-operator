# Cluster Logging Operator Troubleshooting

## Known issues

## Frequently Asked Questions (FAQs)

### 1. I've made changes to my `ClusterLogForwarder` (CLF) instance but the collectors are not redeployed/updated. Why is that?
    
   - There could be an issue with one of the inputs, outputs, or pipelines of the CLF. Check the CLF status in one of 2 ways:
        1. Streamed events of the CLF.
            
            ```$ oc describe clf --show-events=true```

        2. Check the `status` section of the CLF instance `YAML` output.
        
            ```$ oc get clf -oyaml```

### 2. Warning Message: `Currently ignoring file too small to fingerprint`
If you see this warning message in your logs, here is an explanation of what it means and what actions you should take.
This message appears because the file, at the moment it was checked, did not contain a complete line of text.
The default strategy for identifying files is to read the first complete line. A "complete line" is a sequence of text 
that must end with a newline character (\n). This is designed to prevent the system from accidentally processing a log 
entry that is still being written to disk.

### Common Scenario: Container Logs and `conmon`
In container environments that use runtime like CRI-O, a utility called `conmon` manages container I/O.
A key behavior of `conmon` is that it immediately creates an empty log file the moment a container starts, even before 
the application inside the container has written its first log message.
Because the log file exists but is empty (0 bytes), the log collector will find it, see that it doesn't contain a complete line,
and expected to generate this warning.

#### Is This an Error?
Usually, no. In most cases, this is normal and expected behavior.
It typically happens when a file is being created or/and actively written to. An application writes a burst of text but 
has not yet finished the line with a newline character. The system checks the file at this exact moment, sees the incomplete line,
logs the warning, and simply waits to check again. On the next check, the line will likely be complete, and the message will disappear.
This warning might indicate a problem only if it persists for a very long time for a file that you know is no longer being written to.
This could suggest that the application generating the log is not correctly terminating its lines.
#### What to Do
**Observe**: For an actively changing log file, see if the warning appears once, no action is needed. Your system is working perfectly.

**Check the file content**: If the warning persists, open the file mentioned in the message. Look at the very last line.
Does it have a newline character at the end? If not, the application writing the log is not properly terminating its lines.

**Check the source application**: Ensure that the application generating the logs is configured to append a newline character 
(\n) to every log entry. This is standard practice for most logging libraries and systems.