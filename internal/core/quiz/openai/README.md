# OpenAI adapter

The adapter owns credentials, response envelopes, HTTP requests and bounded body
reads. Quiz generation requires an explicit Host for original retained-generation
admission, output validation and trace recording. The host keeps contexts on calls.
The structured transport input is already admitted by its caller; its original
body limit remains distinct from the quiz limit. The shared model package owns
provider-neutral payload and output semantics.
