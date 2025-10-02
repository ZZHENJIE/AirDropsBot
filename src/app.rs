use reqwest::blocking::Client;
use rusqlite::Connection;
use serde::{Deserialize, Serialize};
use std::{fs, io::Read};
use tracing::{Level, event};

#[derive(Debug, Serialize, Deserialize)]
pub struct Email {
    pub smtp_code: String,
    pub smtp_code_type: String,
    pub smtp_email: String,
    pub cola_key: String,
    pub tomail: Vec<String>,
}

#[derive(Debug, Serialize, Deserialize)]
pub struct Config {
    pub port: u16,
    pub log: String,
    pub database: String,
    pub interval: u64,
    pub email: Email,
}

pub struct App {
    pub databese: Connection,
    pub config: Config,
}

impl App {
    pub fn new(args: &Vec<String>) -> Option<Self> {
        match args.get(1) {
            Some(path) => {
                let file = fs::File::open(path);
                match file {
                    Ok(mut file) => {
                        let mut buf = String::new();
                        let _ = file.read_to_string(&mut buf);
                        match serde_json::from_str::<self::Config>(&buf) {
                            Ok(config) => match Connection::open(&config.database) {
                                Ok(databese) => Some(App { databese, config }),
                                Err(e) => {
                                    eprintln!("Failed to open database: {}", e);
                                    None
                                }
                            },
                            Err(e) => {
                                eprintln!("Failed to parse JSON: {}", e);
                                None
                            }
                        }
                    }
                    Err(e) => {
                        eprintln!("Failed to open file: {}", e);
                        None
                    }
                }
            }
            None => {
                eprintln!("Please provide a configuration file");
                None
            }
        }
    }
}

impl Email {
    #[allow(non_snake_case)]
    pub fn sned(
        config: &Config,
        fromTitle: &str,
        subject: &str,
        content: &str,
        isTextContent: bool,
    ) -> Result<bool, reqwest::Error> {
        let client = Client::new();
        for email in config.email.tomail.iter() {
            let body = serde_json::json!({
                "ColaKey":config.email.cola_key,
                "fromTitle": fromTitle,
                "subject": subject,
                "content": content,
                "isTextContent": isTextContent,
                "tomail": email,
                "smtpCode": config.email.smtp_code,
                "smtpEmail": config.email.smtp_email,
                "smtpCodeType": config.email.smtp_code_type,
            });
            let response = client
                .post("https://luckycola.com.cn/tools/customMail")
                .header("Content-Type", "application/json")
                .body(body.to_string())
                .send()?;
            match response.json::<serde_json::Value>() {
                Ok(json) => {
                    if json["code"] != 0 {
                        event!(Level::ERROR, "Email send failed: {}", json["msg"]);
                    }
                }
                Err(e) => event!(Level::ERROR, "Email send failed: {}", e),
            }
        }
        Ok(true)
    }
}
