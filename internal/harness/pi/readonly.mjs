// Only this host-owned extension is loaded for the explicit file inspection profile.
import * as fs from "node:fs";
import * as path from "node:path";
import { TextDecoder } from "node:util";

const config = JSON.parse(fs.readFileSync(path.join(path.dirname(process.env.PI_CODING_AGENT_DIR), "inspection.json"), "utf8"));
const byteLimit = 65536;
const entryLimit = 4096;
const decoder = new TextDecoder("utf-8", { fatal: true });

function guard(target) {
  const relative = path.relative(config.root, target);
  const outside = relative === ".." || relative.startsWith("../") || path.isAbsolute(relative);
  if (config.blocked.includes(target) || (outside && !config.allowed.includes(target))) {
    throw new Error("Path is outside the host file inspection capability.");
  }
  return target;
}

function opened(target, directory, inspect) {
  const resolved = guard(fs.realpathSync(target));
  const flags = fs.constants.O_RDONLY | fs.constants.O_NOFOLLOW | fs.constants.O_NONBLOCK | (directory ? fs.constants.O_DIRECTORY : 0);
  const fd = fs.openSync(resolved, flags);
  try {
    guard(fs.realpathSync(`/proc/self/fd/${fd}`));
    const info = fs.fstatSync(fd);
    if (directory ? !info.isDirectory() : !info.isFile() || info.nlink !== 1 || info.size > byteLimit) {
      throw new Error("Inspection requires a directory or a regular text file within 64 KiB.");
    }
    return inspect(fd);
  } finally {
    fs.closeSync(fd);
  }
}

function read(target) {
  return opened(target, false, fd => {
    // Read at most one byte beyond the ceiling even if the file grows after stat.
    const buffer = Buffer.alloc(byteLimit + 1);
    let size = 0;
    while (size < buffer.length) {
      const count = fs.readSync(fd, buffer, size, buffer.length - size, null);
      if (count === 0) break;
      size += count;
    }
    if (size > byteLimit) throw new Error("File exceeds the inspection byte ceiling.");
    const text = decoder.decode(buffer.subarray(0, size));
    if (text.includes("\0")) throw new Error("Binary file inspection is unsupported.");
    return { text };
  });
}

function list(target, recursive, query, signal) {
  const pending = [target];
  const entries = [];
  let scanned = 0;
  while (pending.length) {
    if (signal?.aborted) throw new Error("File inspection cancelled.");
    const current = pending.pop();
    opened(current, true, fd => {
      const directory = fs.opendirSync(`/proc/self/fd/${fd}`);
      try {
        for (let entry; (entry = directory.readSync()) !== null;) {
          if (++scanned > entryLimit) throw new Error("Directory inspection exceeds 4096 entries; narrow the path.");
          if (entry.isSymbolicLink() || entry.name === ".git") continue;
          const full = path.join(current, entry.name);
          if (config.blocked.includes(full)) continue;
          const relative = path.relative(config.root, full);
          if (!query || relative.toLowerCase().includes(query.toLowerCase())) {
            entries.push({ path: relative, type: entry.isDirectory() ? "directory" : "file" });
          }
          if (recursive && entry.isDirectory()) pending.push(full);
        }
      } finally {
        directory.closeSync();
      }
    });
  }
  entries.sort((a, b) => a.path.localeCompare(b.path));
  return { entries };
}

export default function (pi) {
  pi.registerTool({
    name: "chr_inspect",
    label: "Inspect project files",
    description: "Read-only project inspection. list lists one directory; find searches paths recursively using a literal case-insensitive query; read returns one UTF-8 file. Paths are relative to the host project root. read also accepts exact host-provided evidence file references. No shell, writes, private selected resources or other outside paths. At most 4096 entries and 64 KiB per result; narrow the path when a ceiling is reached.",
    parameters: {
      type: "object", additionalProperties: false, required: ["action", "path"],
      properties: {
        action: { type: "string", enum: ["list", "find", "read"] },
        path: { type: "string", minLength: 1, maxLength: 4096 },
        query: { type: "string", maxLength: 128 },
      },
    },
    async execute(_id, args, signal) {
      try {
        if (signal?.aborted) throw new Error("cancelled");
        const target = path.resolve(config.root, args.path);
        let result;
        if (args.action === "read") result = read(target);
        else if (args.action === "list" || args.action === "find") result = list(target, args.action === "find", args.query, signal);
        else throw new Error("unsupported inspection action");
        const text = JSON.stringify(result);
        if (Buffer.byteLength(text) > byteLimit) throw new Error("result too large");
        return { content: [{ type: "text", text }], details: undefined };
      } catch {
        throw new Error("Read-only inspection failed: use an accessible project path and remain within 4096 entries and 64 KiB.");
      }
    },
  });
}
