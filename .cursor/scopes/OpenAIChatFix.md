# OpenAI Chat Parse Fix

## Purpose
Fix the backend failure when calling the OpenAI-compatible chat completion API so `/api/correct` no longer panics while parsing the response.

## User Problem
The current Go backend crashes when the upstream response does not contain `choices[0].message.content` in the exact shape the code expects.

## Observed Issue
- Current endpoint: `/api/correct`
- Current model: `gpt-4o-mini`
- Current upstream URL: `https://api.gptsapi.net/v1/chat/completions`
- Current failure: `interface conversion: interface {} is nil, not []interface {}`

## Success Criteria
- The backend does not panic on malformed or error responses from the upstream LLM API.
- The backend returns a useful HTTP error to the frontend when the upstream request fails.
- Valid model responses are parsed successfully and returned to the frontend.

## Proposed Scope
- Harden `callLLM()` response parsing.
- Check upstream HTTP status codes before parsing success payloads.
- Surface upstream error details when possible.
- Update `/api/correct` success response to return a proper JSON object parsed from the LLM's `content`.

## Technical Considerations
- The upstream service may return an error JSON instead of a success JSON.
- The `content` field might differ in type or be missing.
- JSON decoding errors are currently ignored and should likely be handled.
- Empty `OPENAI_API_KEY` handling may also need a guard.

## Out Of Scope
- Frontend UI redesign.
- Prompt redesign beyond what is needed to keep the existing flow working.
- Migrating to a different backend framework.

## Questions
1. Continue using `https://api.gptsapi.net/v1/chat/completions` (do not switch).
2. Forward a more specific HTTP status from the upstream (instead of always HTTP 500).
3. On success, return a "real" JSON object (parse the LLM JSON from the `content` string) rather than returning raw JSON bytes as a string.
