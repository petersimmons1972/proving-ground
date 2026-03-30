import click


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
    default="/data",
    show_default=True,
    help="Data directory for profiles and results",
)
def main(tier, data_dir):
    """Proving Ground — AI agent personality benchmark.

    Measures whether agent personality profiles improve task execution
    quality across correctness, elegance, discipline, judgment,
    creativity, and recovery.
    """
    click.echo(f"Proving Ground — running tier={tier}, data={data_dir}")
