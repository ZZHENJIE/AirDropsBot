use crate::{
    Task,
    app::{self, Email},
};
use serde::{Deserialize, Serialize};
use tracing::{Level, event};

const TEN_MINUTES_MS: u64 = 10 * 60 * 1000;
const FIVE_MINUTES_MS: u64 = 5 * 60 * 1000;
const ONE_MINUTE_MS: u64 = 1 * 60 * 1000;

#[allow(non_snake_case)]
#[derive(Debug, Serialize, Deserialize)]
struct AirdropItem {
    configId: String,
    configName: String,
    configDescription: String,
    configSequence: i32,
    pointsThreshold: f32,
    deductPoints: f32,
    binanceChainId: String,
    chainIconUrl: String,
    contractAddress: String,
    tokenSymbol: String,
    tokenLogo: String,
    alphaId: String,
    airdropAmount: f32,
    displayStartTime: u64,
    claimStartTime: u64,
    claimEndTime: u64,
    twoStageFlag: bool,
    stageMinutes: i32,
    secondPointsThreshold: f32,
    claimedRatio: String,
    status: String,
    projectType: String,
    claimInfo: Option<serde_json::Value>,
    claimPreCheckInfo: Option<serde_json::Value>,
    isDelay: bool,
}

#[allow(non_snake_case)]
#[derive(Debug, Serialize, Deserialize)]
struct Token {
    iconUrl: String,
    cnDescription: String,
    enDescription: String,
    price: String,
}

impl Token {
    #[allow(non_snake_case)]
    fn get_info(
        contractAddress: &str,
        chainId: &str,
        clinet: &reqwest::blocking::Client,
    ) -> Result<Self, String> {
        let url = format!(
            "https://www.maxweb.systems/bapi/defi/v1/public/wallet-direct/buw/wallet/cex/alpha/token/full/info?chainId={}&contractAddress={}",
            chainId, contractAddress
        );
        let response = match clinet.get(url).send() {
            Ok(response) => response,
            Err(e) => return Err(e.to_string()),
        };
        match response.json::<serde_json::Value>() {
            Ok(json) => {
                let iconUrl = json["data"]["metaInfo"]["iconUrl"].to_string();
                let cnDescription = json["data"]["metaInfo"]["cnDescription"].to_string();
                let enDescription = json["data"]["metaInfo"]["enDescription"].to_string();
                let price = json["data"]["priceInfo"]["price"].to_string();

                Ok(Token {
                    iconUrl,
                    cnDescription,
                    enDescription,
                    price,
                })
            }
            Err(e) => Err(e.to_string()),
        }
    }
}

#[derive(Debug)]
pub struct Airdrops {
    url: &'static str,
    body: &'static str,
    client: reqwest::blocking::Client,
}

impl Airdrops {
    pub fn default() -> Self {
        Airdrops {
            url: "https://www.binance.com/bapi/defi/v1/friendly/wallet-direct/buw/growth/query-alpha-airdrop",
            body: r#"{"page":1,"rows":20}"#,
            client: reqwest::blocking::Client::new(),
        }
    }
    fn get_data(&self) -> Result<Vec<AirdropItem>, String> {
        match self
            .client
            .post(self.url)
            .header("Content-Type", "application/json")
            .body(self.body)
            .send()
        {
            Ok(response) => match response.json::<serde_json::Value>() {
                Ok(json) => match serde_json::from_value(json["data"]["configs"].clone()) {
                    Ok(items) => Ok(items),
                    Err(e) => Err(e.to_string()),
                },
                Err(e) => Err(e.to_string()),
            },
            Err(e) => Err(e.to_string()),
        }
    }
}

impl Task for Airdrops {
    fn run(&mut self, app: &app::App) -> Result<(), String> {
        let items: Vec<AirdropItem> = match self.get_data() {
            Ok(items) => items,
            Err(e) => {
                event!(Level::ERROR, "Failed to parse configs: {}", e);
                return Err(e.to_string());
            }
        };

        match AirdropsDatabase::init(&app.databese) {
            Ok(_) => {}
            Err(e) => {
                event!(Level::ERROR, "Failed to init database: {}", e);
                return Err(e.to_string());
            }
        }

        let timestamp_ms = match std::time::SystemTime::now().duration_since(std::time::UNIX_EPOCH)
        {
            Ok(t) => t.as_secs() * 1000,
            Err(e) => {
                event!(Level::ERROR, "Failed to get timestamp: {}", e);
                return Err(e.to_string());
            }
        };

        for item in items {
            let token_info =
                match Token::get_info(&item.contractAddress, &item.binanceChainId, &self.client) {
                    Ok(info) => info,
                    Err(e) => {
                        event!(Level::ERROR, "Failed to get token info: {}", e);
                        return Err(e.to_string());
                    }
                };
            let mut db = match AirdropsDatabase::query(&app.databese, &item.contractAddress) {
                Ok(row) => match row {
                    Some(row) => row,
                    None => {
                        let new_row = AirdropsDatabase::new(
                            item.contractAddress.clone(),
                            item.binanceChainId.clone(),
                            item.claimEndTime,
                        );
                        match AirdropsDatabase::insert_or_replace(&app.databese, &new_row) {
                            Ok(_) => new_row,
                            Err(e) => {
                                event!(Level::ERROR, "Failed to insert or replace database: {}", e);
                                return Err(e.to_string());
                            }
                        }
                    }
                },
                Err(e) => {
                    event!(Level::ERROR, "Failed to query database: {}", e);
                    return Err(e.to_string());
                }
            };

            if timestamp_ms > db.claimEndTime {
                continue;
            }

            let mut is_revise = false;
            let remaining_time = db.claimEndTime - timestamp_ms;
            if remaining_time <= TEN_MINUTES_MS && db.tenMinuteReminder == 0 {
                db.tenMinuteReminder = 1;
                is_revise = true;
                match Email::sned(
                    &app.config,
                    "Binance Airdrop",
                    "Binance Airdrop 10 Minutes Reminder",
                    &token_info.cnDescription,
                    true,
                ) {
                    Ok(_) => {}
                    Err(e) => {
                        event!(Level::ERROR, "Failed to send email: {}", e);
                        return Err(e.to_string());
                    }
                }
            }
            if remaining_time <= FIVE_MINUTES_MS && db.fiveMinuteReminder == 0 {
                db.fiveMinuteReminder = 1;
                is_revise = true;
                match Email::sned(
                    &app.config,
                    "Binance Airdrop",
                    "Binance Airdrop 5 Minutes Reminder",
                    &token_info.cnDescription,
                    true,
                ) {
                    Ok(_) => {}
                    Err(e) => {
                        event!(Level::ERROR, "Failed to send email: {}", e);
                        return Err(e.to_string());
                    }
                }
            }
            if remaining_time <= ONE_MINUTE_MS && db.oneMinuteReminder == 0 {
                db.oneMinuteReminder = 1;
                is_revise = true;
                match Email::sned(
                    &app.config,
                    "Binance Airdrop",
                    "Binance Airdrop 1 Minutes Reminder",
                    &token_info.cnDescription,
                    true,
                ) {
                    Ok(_) => {}
                    Err(e) => {
                        event!(Level::ERROR, "Failed to send email: {}", e);
                        return Err(e.to_string());
                    }
                }
            }

            if is_revise {
                match AirdropsDatabase::insert_or_replace(&app.databese, &db) {
                    Ok(_) => {}
                    Err(e) => {
                        event!(Level::ERROR, "Failed to insert or replace database: {}", e);
                        return Err(e.to_string());
                    }
                }
            }
        }
        Ok(())
    }
}

#[allow(non_snake_case)]
struct AirdropsDatabase {
    contractAddress: String,
    binanceChainId: String,
    claimEndTime: u64,
    tenMinuteReminder: u64,
    fiveMinuteReminder: u64,
    oneMinuteReminder: u64,
}

#[allow(non_snake_case)]
impl AirdropsDatabase {
    fn new(contractAddress: String, binanceChainId: String, claimEndTime: u64) -> Self {
        AirdropsDatabase {
            contractAddress: contractAddress,
            binanceChainId: binanceChainId,
            claimEndTime,
            tenMinuteReminder: 0,
            fiveMinuteReminder: 0,
            oneMinuteReminder: 0,
        }
    }
    fn init(connection: &rusqlite::Connection) -> Result<usize, rusqlite::Error> {
        connection.execute(
            "
                CREATE TABLE IF NOT EXISTS binance_airdrops_status (
                    contractAddress TEXT PRIMARY KEY,
                    binanceChainId VERCHAR(3),
                    claimEndTime INT,
                    tenMinuteReminder INT,
                    fiveMinuteReminder INT,
                    oneMinuteReminder INT
                )
        ",
            (),
        )
    }
    fn insert_or_replace(
        connection: &rusqlite::Connection,
        item: &AirdropsDatabase,
    ) -> Result<usize, rusqlite::Error> {
        connection.execute(
        "
        INSERT OR REPLACE INTO binance_airdrops_status 
        (contractAddress, binanceChainId, claimEndTime, tenMinuteReminder, fiveMinuteReminder, oneMinuteReminder)
        VALUES (?, ?, ?, ?, ?, ?)
        ",
        (
            item.contractAddress.clone(),
            item.binanceChainId.clone(),
            item.claimEndTime,
            item.tenMinuteReminder,
            item.fiveMinuteReminder,
            item.oneMinuteReminder,
        ),
    )
    }
    fn query(
        connection: &rusqlite::Connection,
        contractAddress: &str,
    ) -> Result<Option<AirdropsDatabase>, rusqlite::Error> {
        let mut stmt = connection
            .prepare("SELECT * FROM binance_airdrops_status WHERE contractAddress = ?")?;
        stmt.query_row([contractAddress], |row| {
            Ok(AirdropsDatabase {
                contractAddress: row.get(0)?,
                binanceChainId: row.get(1)?,
                claimEndTime: row.get(2)?,
                tenMinuteReminder: row.get(3)?,
                fiveMinuteReminder: row.get(4)?,
                oneMinuteReminder: row.get(5)?,
            })
        })
        .map(Some)
        .or_else(|e| match e {
            rusqlite::Error::QueryReturnedNoRows => Ok(None),
            e => Err(e),
        })
    }
}
