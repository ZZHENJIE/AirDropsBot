use adb::{Task, app, binance};
use std::{thread::sleep, time::Duration};

fn main() {
    let args: Vec<String> = std::env::args().collect();
    match app::App::new(&args) {
        Some(app) => {
            init_tracing(&app.config.log);
            let mut tasks: Vec<Box<dyn Task>> = vec![Box::new(binance::Airdrops::default())];
            loop {
                for task in &mut tasks {
                    match task.run(&app) {
                        Ok(_) => {}
                        Err(e) => {
                            app::Email::sned(&app.config, "Server Error", "Error", &e, true)
                                .unwrap();
                            break;
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
