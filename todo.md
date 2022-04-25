# Add slack message reaction

```
slack-message-timestamp: ${{ fromJson(steps.send-message.outputs.slack-result).response.message.ts }}
```

The timestamp is how you find the right message to reply to.
