from src.tasks import load_tasks, Task


def test_load_tier1_tasks():
    tasks = load_tasks(tiers=["1"])
    assert len(tasks) == 3


def test_task_has_required_fields():
    tasks = load_tasks(tiers=["1"])
    t = tasks[0]
    assert t.id
    assert t.tier in (1, 2, 3)
    assert t.title
    assert t.spec  # full markdown content


def test_tasks_ordered_by_id():
    tasks = load_tasks(tiers=["1"])
    ids = [t.id for t in tasks]
    assert ids == sorted(ids)


def test_load_all_tasks():
    tasks = load_tasks(tiers=["1", "2", "3"])
    assert len(tasks) == 10
