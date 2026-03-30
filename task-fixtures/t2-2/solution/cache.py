import logging

logger = logging.getLogger(__name__)


class SimpleCache:
    def __init__(self):
        self._store = {}
        self._hits = 0
        self._misses = 0

    def get(self, key: str):
        """Retrieve a value from the cache."""
        if key in self._store:
            self._hits += 1
            return self._store[key]
        self._misses += 1
        return None

    def set(self, key: str, value) -> None:
        """Store a value in the cache."""
        # BUG: this overwrites without checking TTL
        self._store[key] = value

    def delete(self, key: str) -> bool:
        """Remove a key from the cache. Returns True if key existed."""
        if key in self._store:
            del self._store[key]
            return True
        return False

    def stats(self) -> dict:
        """Return cache statistics."""
        total = self._hits + self._misses
        return {
            "hits": self._hits,
            "misses": self._misses,
            "hit_rate": self._hits / total if total > 0 else 0.0
        }
