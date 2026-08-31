import fs from "node:fs";
import path from "node:path";
import { spawnSync } from "node:child_process";
import { fileURLToPath } from "node:url";

const repoRoot = path.resolve(
    path.dirname(fileURLToPath(import.meta.url)),
    "..",
);
const errors = [];

const requiredAgentFiles = [
    "AGENTS.md",
    "library/AGENTS.md",
    "scripts/AGENTS.md",
    "src/AGENTS.md",
    "src/cmd/AGENTS.md",
    "src/internal/AGENTS.md",
    "src/internal/config/AGENTS.md",
    "src/internal/handler/AGENTS.md",
    "src/internal/middleware/AGENTS.md",
    "src/internal/model/AGENTS.md",
    "src/internal/service/AGENTS.md",
    "src/internal/store/AGENTS.md",
    "src/migrations/AGENTS.md",
    "vue/AGENTS.md",
    "vue/src/AGENTS.md",
    "vue/src/api/AGENTS.md",
    "vue/src/components/AGENTS.md",
    "vue/src/router/AGENTS.md",
    "vue/src/stores/AGENTS.md",
    "vue/src/views/AGENTS.md",
    "docs/AGENTS.md",
];

function absolute(relativePath) {
    return path.join(repoRoot, relativePath);
}

function read(relativePath) {
    const filePath = absolute(relativePath);
    if (!fs.existsSync(filePath)) {
        errors.push(`missing required file: ${relativePath}`);
        return "";
    }
    return fs.readFileSync(filePath, "utf8");
}

function checkBudget(relativePath, text) {
    const limit = relativePath === "AGENTS.md" ? 2500 : 1400;
    if ([...text].length > limit) {
        errors.push(`${relativePath} exceeds its ${limit}-character budget`);
    }
}

function checkRelativeLinks(relativePath, text) {
    const linkPattern = /\[[^\]]*\]\(([^)]+)\)/g;
    for (const match of text.matchAll(linkPattern)) {
        const target = match[1].trim().replace(/^<|>$/g, "").split("#")[0];
        if (!target || /^[a-z]+:/i.test(target)) continue;
        const resolved = path.resolve(
            path.dirname(absolute(relativePath)),
            target,
        );
        if (!fs.existsSync(resolved)) {
            errors.push(`${relativePath} has a broken link: ${match[1]}`);
        }
    }
}

const skippedDirectories = new Set([
    ".git",
    "dist",
    "node_modules",
    "vue/.pw-browsers",
]);
const agentFiles = [];
function collectAgentFiles(directory) {
    const relativeDirectory = path.relative(repoRoot, directory);
    if (
        skippedDirectories.has(relativeDirectory) ||
        skippedDirectories.has(path.basename(directory))
    ) {
        return;
    }
    for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
        const entryPath = path.join(directory, entry.name);
        if (entry.isDirectory()) collectAgentFiles(entryPath);
        else if (entry.name === "AGENTS.md")
            agentFiles.push(path.relative(repoRoot, entryPath));
    }
}
collectAgentFiles(repoRoot);
agentFiles.sort();

for (const relativePath of requiredAgentFiles) {
    if (!agentFiles.includes(relativePath))
        errors.push(`missing required file: ${relativePath}`);
}

for (const relativePath of agentFiles) {
    const text = read(relativePath);
    if (!text) continue;
    checkBudget(relativePath, text);
    checkRelativeLinks(relativePath, text);
}

const rootAgents = read("AGENTS.md");
for (const requiredLink of ["docs/testing.md", "docs/adr/README.md"]) {
    if (!rootAgents.includes(requiredLink)) {
        errors.push(`AGENTS.md must link to ${requiredLink}`);
    }
}

const adrRoot = absolute("docs/adr");
const adrFiles = [];
function collectAdrFiles(directory) {
    for (const entry of fs.readdirSync(directory, { withFileTypes: true })) {
        const entryPath = path.join(directory, entry.name);
        if (entry.isDirectory()) collectAdrFiles(entryPath);
        else if (entry.name.endsWith(".md"))
            adrFiles.push(path.relative(repoRoot, entryPath));
    }
}
collectAdrFiles(adrRoot);

function checkNotIgnored(relativePath) {
    const result = spawnSync(
        "git",
        ["check-ignore", "-q", "--", relativePath],
        { cwd: repoRoot },
    );
    if (result.status === 0) errors.push(`${relativePath} is ignored by Git`);
    else if (result.status !== 1)
        errors.push(`could not inspect Git ignore state for ${relativePath}`);
}

for (const relativePath of [...agentFiles, ...adrFiles])
    checkNotIgnored(relativePath);

const decisionFiles = adrFiles.filter(
    (relativePath) =>
        !relativePath.endsWith("/README.md") &&
        !relativePath.endsWith("/TEMPLATE.md"),
);
if (decisionFiles.length === 0)
    errors.push("docs/adr must contain at least one decision record");

const requiredSections = {
    proposed: [
        "Problem",
        "Proposal",
        "Alternatives considered",
        "Acceptance criteria",
        "Risks",
    ],
    implemented: [
        "Problem",
        "Decision",
        "Alternatives considered",
        "Consequences",
    ],
    rejected: [
        "Problem",
        "Decision",
        "Alternatives considered",
        "Consequences",
    ],
    archived: [
        "Problem",
        "Decision",
        "Alternatives considered",
        "Consequences",
    ],
};
const allowedClasses = new Set([
    "architecture",
    "process",
    "testing",
    "security",
    "simplification",
]);

for (const relativePath of decisionFiles) {
    const match = relativePath.match(
        /^docs\/adr\/(proposed|implemented|rejected|archived)\/([^/]+)\/(\d{4}-\d{2}-\d{2}-[a-z0-9-]+\.md)$/,
    );
    if (!match) {
        errors.push(`${relativePath} does not match the ADR lifecycle path`);
        continue;
    }

    const [, status, recordClass] = match;
    if (!allowedClasses.has(recordClass)) {
        errors.push(
            `${relativePath} uses unsupported ADR class: ${recordClass}`,
        );
    }

    const text = read(relativePath);
    const lines = text.split(/\r?\n/);
    if (!lines[0].startsWith("# ADR: ") || lines[1] !== `Status: ${status}`) {
        errors.push(
            `${relativePath} must start with an ADR title and matching status`,
        );
    }
    for (const section of requiredSections[status]) {
        if (!text.includes(`## ${section}`)) {
            errors.push(`${relativePath} is missing section: ${section}`);
        }
    }
    checkRelativeLinks(relativePath, text);
}

for (const relativePath of ["docs/adr/README.md", "docs/adr/TEMPLATE.md"]) {
    checkRelativeLinks(relativePath, read(relativePath));
}

if (errors.length > 0) {
    for (const error of errors)
        console.error(`governance check failed: ${error}`);
    process.exit(1);
}

console.log(
    `agent and ADR checks passed (${agentFiles.length} AGENTS files, ${decisionFiles.length} ADR)`,
);
