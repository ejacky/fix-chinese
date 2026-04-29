# LearnChinese - AI Context

## Project Positioning

LearnChinese is an experiment-stage product for non-native Chinese learners.
It helps users correct Chinese sentences quickly, understand mistakes in simple English, and learn more natural native-like phrasing.

## Why This Project Exists

Many foreign learners can write basic Chinese but are unsure whether their sentence sounds correct or natural.
General AI chat tools can help, but answers are often unstructured and not designed for language learning workflows.
This project focuses on structured, teacher-style feedback for Chinese writing practice.

## Current Stage

- MVP / validation stage
- Currently deployed on Vercel
- May move to a standalone domain later
- Main goal now is demand validation, not feature completeness

## Business Constraint (Project-Level)

This project prioritizes fast user value delivery and monetization validation.
The goal is to find a simple, sustainable path to recurring revenue with minimal operational complexity.

## Execution Principles

- Revenue-first: prioritize features that improve user willingness to pay.
- Speed-first: prefer small experiments and short shipping cycles.
- Simplicity-first: avoid heavy engineering before demand is proven.

## Target Users

- Foreign Chinese learners who want fast sentence correction
- HSK learners improving writing quality
- Users who want to sound natural, not only grammatically correct

## Core User Value

- Instant sentence correction
- Clear English explanation of mistakes
- More natural alternatives used by native speakers
- Low friction trial (quick input, quick result)

## Core Workflow (Business View)

1. User enters a Chinese sentence.
2. System returns:
   - corrected sentence
   - concise English explanation
   - 1-2 natural alternatives
3. User learns and iterates with another sentence.

## Current Success Signals

In this phase, success means proving real user demand:

- Users from acquisition channels actually submit sentences
- A meaningful portion of visitors use the correction flow
- Some users return for repeated usage
- Feedback indicates the correction + explanation format is useful

## Scope (Now)

- Chinese sentence correction
- Simple English explanation
- Natural alternative expressions
- Landing page + basic CTA messaging for trial
- Vietnamese UI + explanation support (auto-detected, no visible toggle)

## Out of Scope (Now)

- Account/login system
- Deep personalization
- Large-scale analytics stack
- Enterprise-level product polish
- Additional languages beyond English and Vietnamese

## AI Collaboration Guide (Token-Efficient)

Read this section first before opening more files.

- If task is about product messaging, user value, or landing-page copy:
  - prioritize `index.html`
- If task is about correction pipeline, API behavior, or response structure:
  - prioritize `api/correct.go`
- If task is about strategy, experiment direction, or what to build next:
  - rely on this `AI.md` first, then ask for missing business decisions

## Non-Negotiable Product Intent

This project is not trying to become a generic chatbot.
It is a focused Chinese-learning assistant that gives structured corrections and learning-oriented explanations.
