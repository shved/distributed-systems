use std::io::StdoutLock;

use serde::{Deserialize, Serialize};
use serde_json;

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Message {
    src: String,
    dest: String,
    body: Body,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
struct Body {
    msg_id: Option<usize>,
    in_reply_to: Option<usize>,
    #[serde(flatten)]
    payload: Payload,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
#[serde(tag = "type")]
#[serde(rename_all = "snake_case")]
enum Payload {
    Echo { echo: String },
    EchoOk { echo: String },
    Init { node_id: String },
    InitOk,
}

enum Error {
    // Timeout(0),  // "Timed out"
    // NodeNotFound(1),  // "Node not found"
    // NotSupported(10), // "Operation not supported"
    // Unavailable(11), // "Temporary unavailable"
    // Malformed(12), // "Malformed request"
    // Crashed(13), // "Crashed"
    // Aborted(14), // "Aborted"
    // KeyNotExist(20), // "Key does not exist"
    // KeyAlreadyExist(21), // "Key already exist"
    // PreconditionFailed(22), // "Precondition failed"
    // TXConflict(30), // "Transaction conflict"
}

struct EchoNode {
    id: String,
    msg_id_seq: usize,
}

impl EchoNode {
    pub fn handle(&mut self, input: Message, output: &mut serde_json::Serializer<StdoutLock>) -> Result<Message, Error> {
        match input.body.payload {
            Payload::Init { node_id, .. } => {
                let reply = Message {
                    src: input.dest, 
                    dest: input.src,
                    body: Body {
                        
                    }
                }
            }
            Payload::Echo { echo, .. } => {
                let reply = Message {
                    src: input.dest,
                    dest: input.src,
                    body: Body {
                        msg_id: Some(self.id),
                        in_reply_to: Some(input.body.id),
                        payload: Payload::EchoOk { echo },
                    },
                };
                self.id += 1;
            }
            Payload::EchoOk { .. } => {}
        }

        Ok(())
    }
}

fn main() -> anyhow::Result<()> {
    let stdin = std::io::stdin().lock();
    let inputs = serde_json::Deserializer::from_reader(stdin).into_iter::<Message>();

    let stdout = std::io::stdout().lock();
    let output = serde_json::Serializer::new(stdout);

    let state = EchoNode {
        id: String::default(),
        msg_id_seq: 0,
    };

    for input in inputs {
        let input = input.context("input could not be deserialized")?;
    }

    Ok(())
}
