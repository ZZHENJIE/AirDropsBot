pub mod app;
pub mod binance;

pub trait Task {
    fn run(&mut self, app: &app::App) -> Result<(), String>;
}
