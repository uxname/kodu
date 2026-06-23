---
name: post-call-task-builder
description: Create structured technical specifications and development tasks from meeting transcripts OR refine and organize user-provided task lists (optionally enriched with transcript context). Use after team meetings, planning sessions, architecture discussions, or when user provides a raw task list that needs structuring, prioritization, and validation against the codebase.
---

# SKILL: Creating and refining tasks

## PURPOSE

This skill works in two modes:

**Mode A — From a call:**

- read the conversation transcript;
- understand the decisions, problems, and agreements;
- study the current project's codebase;
- ask clarifying questions;
- propose a Pareto-optimal solution;
- prepare a clear spec and task list.

**Mode B — From a ready list (with possible context from a call):**

- accept a task list from the user;
- if the user also provided a transcript — extract context, reasons, and priorities from it;
- study the codebase;
- check the tasks for realism, completeness, and coherence;
- refine them: add context, definitions of done, execution order;
- point out risks, dependencies, and hidden complexity.

The main goal is to turn the input (a conversation or a list) into a practical work plan, written in plain language, as
for a junior.

## CRITICAL CONSTRAINTS (MUST NOT BE BROKEN)

**1. Full codebase research is MANDATORY before formulating any tasks.**

- Before writing even a single task, spec, or plan item — you MUST conduct full research of the current project.
- The research includes: project structure, stack, architecture, key modules, entry points, dependencies, tests, configs, README.
- No exceptions. No research → no tasks.

**2. DO NOT ASK the user "what should I analyze?" or "which modules should I look at?".**

- You determine the scope of research yourself and conduct it automatically.
- Your job is to study the entire project thoroughly. The user shouldn't have to point you where to look.
- If the project is large — start from the root (package.json, tsconfig, README, the structure of src/), then go deeper into the modules related to the task topic.

**3. Formulating tasks — ONLY after the research is complete.**

- Start formulating tasks only when you have a full picture of the project.
- Each task must be grounded in the specific files, modules, and architectural decisions you saw in the code.
- If you haven't studied the project — go back to research, don't make up tasks "out of thin air".

## ACTIVATION

Use this skill if the user:

- sent a call transcript or pointed to a file with one;
- asks you to go over the outcomes of a call;
- asks you to compile tasks, a spec, an implementation plan, or a backlog from the results of a discussion;
- provided their own task list and asks you to refine, structure, or check it;
- provided both a task list and a transcript — so the tasks are compiled with the conversation's context in mind;
- wants you to first study the project's codebase and then write up the tasks.

## DO NOT USE WHEN

Don't use this skill if:

- the user doesn't want tasks or a spec;
- you just need to briefly summarize a call without analyzing the project;
- the user asks you to ignore the codebase;
- the user already provided a finished spec and only needs the text edited (without checking against the code).

## INPUTS

**Required (don't start without them):**

- **access to the project's codebase** — you must research the project BEFORE formulating tasks. If the path isn't given — determine it from the working directory context. If the project is unavailable — say so and don't continue.
- the place where the result should be written.

**Additionally for mode A:**

- the path to the transcript file or its text.

**Additionally for mode B:**

- the task list from the user (text, file, message);
- optionally: a transcript or a link to the call for context (helps understand the reasons, priorities, hidden requirements).

**What you DON'T NEED to ask the user:**

- "Which modules should I look at?" — determine it yourself from the project structure.
- "Which files are important for this task?" — find them yourself through research.
- "Where is the relevant functionality?" — trace it yourself through the codebase.
- "What exactly should I analyze?" — analyze EVERYTHING relevant to the task topic.

If you're only missing the transcript (mode A) or the task list (mode B) — ask a short clarifying question. But research the codebase in any case.

## PROCESS

### Step 0. Determine the mode and research the project

**This step is MANDATORY. Don't move on to formulating tasks without it.**

#### 0.1. Determine the mode

- If the user gave only a transcript (no task list) → mode A.
- If the user gave a ready task list (with or without a transcript) → mode B.
- The transcript in mode B is used as additional context — to understand the reasons behind the tasks, their priorities, and expectations more precisely.
- If only a transcript is given but the user explicitly asks "compile tasks" → mode A.

#### 0.2. Conduct full codebase research (MANDATORY)

**Don't ask the user "what should I analyze?" — research the project yourself.**

Research order:

1. **Project root** — read package.json (or equivalent), README, configs (tsconfig, biome, eslint, etc.), .env.example.
2. **Structure** — determine the organization of src/ (or another main directory), the key directories and their purpose.
3. **Entry point** — find the main entry point, understand how the application starts.
4. **Architecture** — determine the pattern (MVC, modular, service layer, etc.), the dependencies between parts.
5. **Key modules** — study the main modules/services, their responsibilities and connections.
6. **Dependencies** — look at which libraries are used and why.
7. **Tests** — whether there are tests, which framework, what coverage is lacking.
8. **Relation to the topic** — go deeper into the parts of the code directly related to the tasks from the transcript or the user's list.

If the project is large — start from the most likely points of change, then broaden the survey. But in the end you must have a full picture.

**The result of the research** — you should be able to answer these questions for yourself:
- How is the project structured overall?
- Which modules/services are relevant to the tasks?
- What constraints exist in the current architecture?
- Where the needed functionality already exists, and what is missing?
- What could break when changes are made?

Only after this should you move on to formulating tasks.

---

### Mode A: From a call

#### A1. Read the transcript

Extract from it:

- the call's goals;
- problems;
- decisions;
- points of contention;
- dependencies;
- risks;
- mentioned files, modules, services, endpoints, tables, queues, integrations.

#### A2. Match the transcript against the research results

You already researched the project in step 0.2. Now match what you saw in the code against what was discussed on the call:

- what already exists in the project on the call's topic;
- what is missing;
- which items from the call can realistically be implemented quickly;
- which require a larger rework;
- where there are technical risks or hidden complexity.

#### A3. Propose a Pareto-optimal solution

Before the clarifying questions, always present the recommended option.

A Pareto-optimal solution is one that:

- gives maximum benefit for minimum complexity;
- captures most of the task's value;
- doesn't require unnecessary architectural restructuring;
- can be implemented faster and more safely than the "ideal" option.

Present it like this:

- what is proposed to do;
- why this is the best balance between cost and benefit;
- what limitations this approach has;
- what will be left for the next stage.

#### A4. Ask clarifying questions

After the analysis, ask only the questions without which the tasks can't be compiled well.

Requirements for the questions:

- ask only the genuinely important questions;
- no more than 3–7 questions at a time if possible;
- the questions must be specific;
- next to each question, give your recommended answer if it's obvious.

Question format:

- the question;
- why it's needed;
- the recommended option, if it's obvious.

#### A5. Compile the spec and tasks

Once the user answers, prepare the result and write it where they specified.

---

### Mode B: From a ready task list

#### B0. If there's a transcript — read it for context

Extract from it:

- the discussion's goals — why these tasks were needed in the first place;
- the problems the tasks should solve;
- priorities and expectations;
- mentioned files, modules, services;
- points of contention — to know where disagreements are possible.

This context will be needed in step B3 when refining each task.

#### B1. Study the task list

Read and classify each task:

- how specific the wording is;
- whether there's a definition of done;
- whether the place to change is specified;
- whether there are hidden dependencies;
- whether there are duplicates or overlaps.

#### B2. Verify the tasks against the research results

You already researched the project in step 0.2. Now check each task from the list against the code:

- whether the task is realistic in the current architecture;
- whether there are already ready-made solutions or analogs;
- which files/modules will need to change;
- what technical constraints exist;
- what side effects might arise.

#### B3. Refine each task

For each task:

- make the wording concrete;
- add context (why it's needed);
- specify exactly where the changes should be made;
- add a definition of done;
- specify dependencies on other tasks;
- note the risks.

If a task is unrealistic or contradicts the code — point this out explicitly and propose an alternative.

#### B4. Assemble the final plan

- sort the tasks in execution order;
- group them by stage (preparation, backend, frontend, tests, docs, deployment);
- add the overall context and goal;
- specify what's out of scope.

Ask clarifying questions only if data is critically lacking (the goal is unclear, there are no priorities, the stack is unclear).

## DECISION RULES

**PRIORITY 0: Mandatory research (ABOVE EVERYTHING)**

- Never formulate tasks without full codebase research.
- Never ask the user "what should I analyze?" — do it yourself.
- If you doubt whether you've studied the project enough — go back and read more.

**PRIORITY 1: Accuracy**

- don't make up details;
- all conclusions rest on the transcript or the code;
- if the transcript has contradictions — point them out and propose the most likely interpretation;
- in mode B the transcript is used only for context — don't rewrite the user's tasks based on the transcript unless they asked for it; refine the original list, don't replace it.

**PRIORITY 2: Practicality**

- propose the minimal change that gives the maximum effect;
- don't propose a solution that breaks the project's current style without need;
- if there are several paths — choose the one that's simpler to implement, easier to maintain, and least likely to break current
  functionality.

**PRIORITY 3: Clarity**

- write as for a junior;
- avoid abstract phrases like "improve the system";
- each task must be specific, verifiable, and achievable.

## TOOL USAGE

**Reading code:**

- use it to find existing functionality, integration points, architectural constraints;
- in step 0.2 — research the project FULLY: root, structure, key modules, dependencies;
- in steps A2/B2 — focus on the specific points related to the tasks;
- note the files a change might affect.

**Code search:**

- use it to check for existing solutions, constants, types, imports;
- in step 0.2 — search for EVERYTHING related to the task topic, even if it seems non-obvious.

**Reading files:**

- before proposing changes, read the existing code at the points of change;
- check whether there are tests for the affected modules.

## OUTPUT REQUIREMENTS

Write in plain, clear language, as for a junior.

Avoid:

- unnecessary theory;
- complex words without need;
- overloaded phrasing;
- abstract phrases like "improve the system" without explanation.

The spec / work plan must include:

- the goal;
- the context;
- the current problem (for mode A) or the original list and the context from the transcript (for mode B);
- the proposed solution / the refined list;
- the execution order;
- the definitions of done;
- the risks and constraints;
- what's out of scope, if that matters.

Where appropriate, break the tasks down by meaning:

- preparation;
- backend;
- frontend;
- integrations;
- testing;
- documentation;
- deployment.

**Task format:**

- name;
- goal;
- what to do;
- where to do it;
- result;
- definition of done;
- dependencies.

## FAILURE HANDLING

**If the codebase is unavailable:**

- don't make up tasks based on assumptions;
- say honestly: "I can't access the project. Without researching the codebase I can't compile tasks.";
- ask the user to specify the path to the project or make sure the directory is accessible.

If there isn't enough data (the transcript is incomplete, the task list is too vague):

- research the codebase IN ANY CASE — it's available regardless of the transcript;
- say honestly what's missing from the transcript/list;
- list what you managed to understand;
- ask the minimal set of clarifying questions;
- don't start writing the spec blindly.

If the transcript or list has contradictions:

- point them out;
- propose the most likely interpretation;
- ask for confirmation.

If a task from the user's list is technically unrealistic:

- explain why;
- propose an alternative that solves the same problem.

## EXAMPLES

### Mode A — Good result:

- Read the transcript: the team discussed speeding up sign-up.
- **Automatically** research the project: find the sign-up module, look at the current flow, understand the architecture, check the tests.
- Match: what already exists (email validation on the frontend, but not on the backend), what's missing (rate limiting, caching).
- Propose a Pareto solution: don't rewrite the whole flow, but add the missing validation and a single service layer.
- Ask whether the old scenarios need to be supported.
- After the answer, write up the spec with tasks for backend, tests, and documentation.

### Mode A — Bad result:

- Ask the user: "Which modules should I look at?" — instead of researching the project yourself.
- Summarize the call in general terms without analyzing the code.
- Propose a large refactor right away without researching the current architecture.
- Write a spec with unclear terms and no definitions of done.

### Mode B (with a list) — Good result:

- The user gave a list: "add email validation, build a profile page, write tests".
- **Automatically** research the project: find where validation already exists, look at the page structure, check test coverage.
- Check against the code: email validation partially exists in sign-up; there's no profile page; there are no validation tests.
- Refine: make the validation concrete (which cases), specify where to create the page, add definitions of done and an
  execution order.
- Result: 3 tasks with context, place of changes, and DoD.

### Mode B (with a list + transcript) — Good result:

- The user gave a list: "add email validation, build a profile page" and a call transcript.
- **Automatically** research the project: find the users module, look at the DB schema, check the endpoints.
- From the transcript: it turned out email validation is needed because of support complaints, and the profile page — for GDPR compliance.
- Check against the code: validation partially exists; there's no profile page; the database doesn't store the consent date for data processing.
- Refine: add a task "add the consent_date field to User", adjust the priorities (GDPR — higher).
- Result: 4 tasks with the correct order and context.

### Mode B — Bad result:

- Ask the user: "Which files should I look at?" — instead of researching automatically.
- Just reword the user's tasks in your own words without checking against the code.
- Don't check against the code whether the tasks are realistic.
- Don't add definitions of done.
- Don't specify order and dependencies.
- Don't use the transcript even though it was provided — important context is missed (GDPR, support priorities).
