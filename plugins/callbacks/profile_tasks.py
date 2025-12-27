# The MIT License (MIT)
#
# Copyright (c) 2014 Jharrod LaFon
#
# Permission is hereby granted, free of charge, to any person obtaining a copy of
# this software and associated documentation files (the "Software"), to deal in
# the Software without restriction, including without limitation the rights to
# use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
# the Software, and to permit persons to whom the Software is furnished to do so,
# subject to the following conditions:
#
# The above copyright notice and this permission notice shall be included in all
# copies or substantial portions of the Software.
#
# THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
# IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
# FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
# COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
# IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
# CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

"""
Ansible callback plugin for timing task execution.

This module follows SOLID principles with clear separation of concerns.
"""
from __future__ import annotations

import time
from dataclasses import dataclass, field
from typing import Dict, List, Optional, Tuple


@dataclass
class TaskTiming:
    """Value object representing a task's timing information."""

    name: str
    start_time: float
    end_time: Optional[float] = None

    @property
    def elapsed_time(self) -> float:
        """Calculate elapsed time for the task."""
        if self.end_time is None:
            return 0.0
        return self.end_time - self.start_time

    @property
    def is_complete(self) -> bool:
        """Check if the task timing is complete."""
        return self.end_time is not None


@dataclass
class TimingStatistics:
    """Domain service for managing task timing statistics."""

    timings: Dict[str, TaskTiming] = field(default_factory=dict)
    max_results: int = 10

    def start_task(self, name: str) -> None:
        """Record the start time of a task."""
        self.timings[name] = TaskTiming(name=name, start_time=time.time())

    def end_task(self, name: str) -> None:
        """Record the end time of a task."""
        if name in self.timings:
            self.timings[name].end_time = time.time()

    def get_slowest_tasks(self) -> List[Tuple[str, float]]:
        """
        Get the slowest tasks.

        Returns:
            List of (task_name, elapsed_time) tuples, sorted by elapsed time descending.
        """
        completed_tasks = [
            (timing.name, timing.elapsed_time)
            for timing in self.timings.values()
            if timing.is_complete
        ]

        return sorted(completed_tasks, key=lambda x: x[1], reverse=True)[: self.max_results]

    def format_results(self) -> str:
        """Format timing results for display."""
        results = self.get_slowest_tasks()

        lines = []
        for name, elapsed in results:
            line = f"{name:-<70}{elapsed:->9.02f}s"
            lines.append(line)

        return "\n".join(lines)


class CallbackModule:
    """
    Ansible callback plugin for profiling task execution times.

    This implementation follows the Single Responsibility Principle by
    delegating statistics management to TimingStatistics.
    """

    def __init__(self):
        """Initialize the callback module."""
        self.stats = TimingStatistics()
        self.current_task: Optional[str] = None

    def playbook_on_task_start(self, name: str, is_conditional: bool) -> None:
        """
        Handle task start event.

        Args:
            name: The task name
            is_conditional: Whether the task is conditional
        """
        # End the previous task if one is running
        if self.current_task is not None:
            self.stats.end_task(self.current_task)

        # Start the new task
        self.current_task = name
        self.stats.start_task(name)

    def playbook_on_stats(self, stats) -> None:
        """
        Handle playbook completion and print timing statistics.

        Args:
            stats: Ansible statistics object
        """
        # End the last task
        if self.current_task is not None:
            self.stats.end_task(self.current_task)

        # Print the results
        print(self.stats.format_results())
