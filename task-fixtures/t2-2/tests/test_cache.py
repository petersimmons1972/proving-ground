import sys
import os
import logging
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from solution.cache import SimpleCache


def test_get_logs_cache_hit(caplog):
    cache = SimpleCache()
    cache.set("key1", "value1")
    with caplog.at_level(logging.DEBUG):
        cache.get("key1")
    assert "Cache hit: key1" in caplog.text


def test_get_does_not_log_on_miss(caplog):
    cache = SimpleCache()
    with caplog.at_level(logging.DEBUG):
        cache.get("nonexistent")
    assert "Cache hit" not in caplog.text


def test_set_method_unchanged():
    cache = SimpleCache()
    cache.set("k", "v")
    assert cache.get("k") == "v"


def test_delete_method_unchanged():
    cache = SimpleCache()
    cache.set("k", "v")
    assert cache.delete("k") is True
    assert cache.get("k") is None


def test_stats_method_unchanged():
    cache = SimpleCache()
    cache.set("k", "v")
    cache.get("k")
    cache.get("missing")
    stats = cache.stats()
    assert stats["hits"] == 1
    assert stats["misses"] == 1
