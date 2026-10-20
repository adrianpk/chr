// The process runs in private state, so project Pi settings cannot be loaded.
// Reuse the pinned SDK's builtin implementations at the host-selected checkout.
import * as fs from "node:fs";
import * as path from "node:path";
import {
  createReadTool, createBashTool, createEditTool, createWriteTool,
  createGrepTool, createFindTool, createLsTool,
} from "/opt/chr-pi/node_modules/@earendil-works/pi-coding-agent/dist/index.js";

const { root } = JSON.parse(fs.readFileSync(path.join(process.cwd(), "worker.json"), "utf8"));
export default function (pi) {
  for (const factory of [createReadTool, createBashTool, createEditTool, createWriteTool, createGrepTool, createFindTool, createLsTool]) {
    const tool = factory(root);
    pi.registerTool({
      name: tool.name, label: tool.label, description: tool.description, parameters: tool.parameters,
      execute(id, args, signal, onUpdate, ctx) {
        return tool.execute(id, args, signal, onUpdate, { ...ctx, cwd: root });
      },
    });
  }
  pi.on("before_agent_start", event => ({
    systemPrompt: event.systemPrompt.replaceAll(process.cwd(), root),
  }));
}
