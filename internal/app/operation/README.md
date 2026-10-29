# Application operation outcomes

`operation` owns transport-neutral result values shared by application adapters.
Results describe status, summary, canonical findings, known changed paths and
eligible recovery. Adapters bind recovery to captured input and present their
own navigation or interaction controls. Domain operations retain ownership of
revision guards, receipts and durable writes.
