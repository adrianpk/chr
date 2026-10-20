// Owned deterministic model endpoint. It uses no account or external network.
import http from "node:http";

let calls = 0;
let released = process.env.CHR_FIXTURE_ISOLATION !== "1";
const pending = [];
const credentials = new Map([
  ["Bearer worker-fixture-only", "single"],
  ["Bearer worker-fixture-first", "first"],
  ["Bearer worker-fixture-second", "second"],
]);
const command = label => [
  "test ! -e /var/run/docker.sock",
  "test -z \"$OPENAI_API_KEY\"",
  `printf 'CALL=${label}\\n' > call.txt`,
  "printf 'Delivered\\nHOME=%s\\nSTATE=%s\\n' \"$HOME\" \"$PI_CODING_AGENT_DIR\" > result.txt",
  "git add result.txt call.txt",
  "git -c user.name=Test -c user.email=test@example.com commit -qm 'Pi task result'",
].join(" && ");

http.createServer(async (req, res) => {
  if (req.url === "/health") { res.end(JSON.stringify({waiting: pending.length, calls})); return; }
  if (req.url === "/release" && req.method === "POST") {
    released = true;
    for (const resolve of pending.splice(0)) resolve();
    res.end("released"); return;
  }
  const chunks = [];
  let size = 0;
  for await (const chunk of req) {
    size += chunk.length;
    if (size > 4194304) { res.writeHead(413).end(); return; }
    chunks.push(chunk);
  }
  const text = Buffer.concat(chunks).toString("utf8");
  const request = JSON.parse(text);
  const tools = request.tools.map(tool => tool.function.name).sort();
  const expected = ["bash", "edit", "find", "grep", "ls", "read", "write"];
  const label = credentials.get(req.headers.authorization);
  const valid = req.url === "/v1/chat/completions" && label &&
    !text.includes("credential_file") && !text.includes("worker-fixture-") &&
    !text.includes("POISON_WORKER_CONTEXT") &&
    (process.env.CHR_FIXTURE_ISOLATION !== "1" || text.includes("Explicit project fixture instructions.")) && JSON.stringify(tools) === JSON.stringify(expected);
  if (!valid) { console.log(JSON.stringify({invalid: true})); res.writeHead(400).end(); return; }
  calls++;
  console.log(JSON.stringify({calls}));
  const second = request.messages.some(message => message.role === "tool");
  if (second) console.log(JSON.stringify({tool_results: request.messages.filter(message => message.role === "tool").map(message => message.content)}));
  if (!second && !released) await new Promise(resolve => pending.push(resolve));
  const delta = second ? {role:"assistant",content:'{"done":true}'} :
    {role:"assistant",tool_calls:[{index:0,id:"worker-write",type:"function",function:{name:"bash",arguments:JSON.stringify({command: command(label)})}}]};
  const chunk = {id:"owned",object:"chat.completion.chunk",model:"model",choices:[{index:0,delta,finish_reason:second?"stop":"tool_calls"}]};
  res.writeHead(200,{"Content-Type":"text/event-stream"});
  res.end(`data: ${JSON.stringify(chunk)}\n\ndata: [DONE]\n\n`);
}).listen(4317,"0.0.0.0");
