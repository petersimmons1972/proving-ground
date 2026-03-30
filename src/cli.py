import click
from pathlib import Path
from src.orchestrator import run_benchmark

_DEFAULT_DATA_DIR = str(Path(__file__).parent.parent / "data")


@click.command()
@click.option(
    "--tier",
    type=click.Choice(["1", "2", "3", "all"]),
    default="all",
    show_default=True,
    help="Run specific tier only",
)
@click.option(
    "--data-dir",
    default=_DEFAULT_DATA_DIR,
    show_default=True,
    help="Data directory for profiles and results",
)
def main(tier: str, data_dir: str) -> None:
    """Proving Ground — AI agent personality benchmark.

    Measures whether agent personality profiles improve task execution
    quality across correctness, elegance, discipline, judgment,
    creativity, and recovery.
    """
    tiers = ["1", "2", "3"] if tier == "all" else [tier]
    click.echo(f"Proving Ground — running tiers={tiers}, data={data_dir}")
    run_benchmark(data_dir=data_dir, tiers=tiers)
    click.echo(f"Done. Results at {data_dir}/results.html")
