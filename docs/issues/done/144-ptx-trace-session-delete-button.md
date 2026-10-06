# Issue #144: PTX-TRACE Session Delete Button

**Type**: Enhancement
**Priority**: Medium
**Status**: ✅ Implemented
**Component**: ptx-trace / server / dashboard
**Labels**: enhancement, ptx-trace, ui, dashboard, session-management

## Summary

Add a delete button to each session item in the PTX-TRACE web dashboard sidebar. The button should be positioned in the top-right corner of each session panel and allow users to delete sessions including all associated trace data.

## Current State

- Dashboard displays sessions in left sidebar (`src/helpers/ptx-trace/server/dashboard.go`)
- Session items show name, status badge, and start time
- Backend already has `DeleteSession()` method in storage (`src/helpers/ptx-trace/storage/storage.go:354`)
- No UI mechanism exists to delete sessions
- No API endpoint exists for session deletion

## Requirements

### Functional Requirements

1. **Delete Button UI**
   - Add delete button (trash/bin icon 🗑) to each session item
   - Position in top-right corner of session panel
   - Button should be visible on hover or always visible
   - Button should have distinct styling (red/danger color)

2. **Confirmation Dialog**
   - Show confirmation dialog before deletion: "Delete session '{name}'? This action cannot be undone."
   - Dialog should have "Cancel" and "Delete" buttons
   - Prevent accidental deletions

3. **API Endpoint**
   - Add `DELETE /api/sessions/:id` endpoint
   - Return 200 OK on success
   - Return 404 if session not found
   - Return 500 on storage error

4. **Session Deletion**
   - Delete session directory including all chunks, index, and manifest
   - Update session list in UI after deletion
   - If deleted session was currently selected, clear the main panel

5. **Real-time Update**
   - Broadcast session deletion event via WebSocket
   - Other connected clients should update their session list

### Non-Functional Requirements

- Delete operation should complete within 2 seconds for typical sessions
- UI should show loading state during deletion
- Error messages should be user-friendly

## Technical Design

### 1. API Endpoint (api.go)

```text
Location: src/helpers/ptx-trace/server/api.go

Add DELETE handler to handleSession():
- Check HTTP method for DELETE
- Call storage.DeleteSession(sessionID)
- Broadcast deletion event to WebSocket clients
- Return appropriate HTTP status
```

### 2. Dashboard UI Changes (dashboard.go)

```text
Location: src/helpers/ptx-trace/server/dashboard.go

CSS additions:
- .session-item { position: relative; }
- .delete-btn styling (positioned absolute top-right)
- .delete-btn:hover styling (red background)

HTML template changes:
- Add delete button inside session-item
- Add confirmation modal HTML

JavaScript additions:
- deleteSession(sessionId) function
- Confirmation dialog handling
- API call with fetch DELETE
- UI update after successful deletion
- WebSocket message handling for deletions from other clients
```

### 3. Storage (already implemented)

```text
Location: src/helpers/ptx-trace/storage/storage.go:354

DeleteSession(sessionID string) - already exists
- Removes entire session directory with os.RemoveAll()
```

## UI Mockup

```text
┌─────────────────────────────────────┐
│ Sessions                            │
├─────────────────────────────────────┤
│ ┌─────────────────────────────────┐ │
│ │ ETL Pipeline Run          [×]   │ │  <- Delete button in top-right
│ │ ● Active  10:45:32              │ │
│ └─────────────────────────────────┘ │
│ ┌─────────────────────────────────┐ │
│ │ API Integration Test      [×]   │ │
│ │ ○ Completed  09:12:15           │ │
│ └─────────────────────────────────┘ │
└─────────────────────────────────────┘
```

## Acceptance Criteria

- [ ] Delete button visible on each session item in sidebar
- [ ] Confirmation dialog appears before deletion
- [ ] Session and all data removed from filesystem after deletion
- [ ] Session list updates after deletion (removed from UI)
- [ ] If active session deleted, main panel shows empty state
- [ ] Other connected WebSocket clients receive deletion notification
- [ ] Error handling for failed deletions (show error message)
- [ ] No data loss for other sessions

## Files to Modify

| File | Changes |
| ---- | ------- |
| `src/helpers/ptx-trace/server/api.go` | Add DELETE handler in handleSession() |
| `src/helpers/ptx-trace/server/dashboard.go` | Add CSS, HTML, and JavaScript for delete button |

## Testing

1. **Manual Testing**
   - Start ptx-trace server
   - Create multiple test sessions
   - Verify delete button appears on each session
   - Click delete, verify confirmation dialog
   - Confirm deletion, verify session removed from list
   - Verify session directory deleted from filesystem
   - Test cancellation of delete operation
   - Test deletion of currently selected session

2. **Edge Cases**
   - Delete session with large amount of data
   - Delete session while events are being written
   - Delete non-existent session (race condition)
   - Multiple clients - delete in one, verify update in another

## Related Issues

- #141 - PTX-TRACE Universal Tracing Helper (parent feature)

## Notes

- Backend `DeleteSession()` already implemented and tested
- Consider adding "Delete All Sessions" button in future enhancement
- Consider adding session archiving as alternative to deletion
