use adb::{Task, app, binance};
use std::{thread::sleep, time::Duration};

fn main() {
    let args: Vec<String> = std::env::args().collect();
    match app::App::new(&args) {
        Some(app) => {
            init_tracing(&app.config.log);
            let mut error_timestamp = std::time::SystemTime::now()
                .duration_since(std::time::UNIX_EPOCH)
                .unwrap()
                .as_secs();
            let mut tasks: Vec<Box<dyn Task>> = vec![Box::new(binance::Airdrops::default())];
            loop {
                for task in &mut tasks {
                    match task.run(&app) {
                        Ok(_) => {}
                        Err(e) => {
                            let error_temp_timestamp = std::time::SystemTime::now()
                                .duration_since(std::time::UNIX_EPOCH)
                                .unwrap()
                                .as_secs();
                            if error_temp_timestamp - error_timestamp > 120 {
                                app::Email::sned(&app.config, "Server Error", "Error", &e, true)
                                    .unwrap();
                                error_timestamp = error_temp_timestamp;
                            }
                        }
                    }
                }
                sleep(Duration::from_secs(app.config.interval));
            }
        }
        None => {}
    }
}

fn init_tracing(log: &str) {
    let file = std::fs::File::create(log).unwrap();
    tracing_subscriber::fmt().with_writer(file).init();
}
