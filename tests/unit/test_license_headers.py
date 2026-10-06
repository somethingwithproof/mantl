# SPDX-License-Identifier: Apache-2.0
"""License checks preserve source bytes, upstream declarations and explicit scope."""


import pytest

from ci import check_license_headers as licensing


def fixture(tmp_path, metadata=""):
    (tmp_path / "LICENSES").mkdir()
    for name in ("Apache-2.0", "MIT", "AGPL-3.0-or-later"):
        (tmp_path / "LICENSES" / f"{name}.txt").write_text("fixture license text")
    (tmp_path / "REUSE.toml").write_text("version = 1\n" + metadata)
    return tmp_path


@pytest.mark.parametrize(
    "header",
    [
        "// SPDX-License-Identifier: MIT",
        "# SPDX-License-Identifier: MIT",
        "/* SPDX-License-Identifier: MIT */",
        "<!-- SPDX-License-Identifier: MIT -->",
        ".. SPDX-License-Identifier: MIT",
        "{{/* SPDX-License-Identifier: MIT */ -}}",
        "SPDX-License-Identifier: MIT",
    ],
)
def test_supported_comment_forms(header, tmp_path):
    file = tmp_path / "source"
    file.write_text("unrelated first line\n" + header + "\nbody")
    assert licensing.header_identifier(file) == "MIT"


def test_missing_unknown_and_empty_declarations_fail(tmp_path):
    root = fixture(tmp_path)
    for name, content in {
        "missing.py": "print('fixture')",
        "unknown.py": "# SPDX-License-Identifier: Unregistered-License",
        "empty.py": "# SPDX-License-Identifier:",
    }.items():
        (root / name).write_text(content)
    assert licensing.check(root, ["missing.py", "unknown.py", "empty.py"]) == [
        "missing.py",
        "unknown.py",
        "empty.py",
    ]


def test_companion_keeps_json_and_binary_bytes_unchanged(tmp_path):
    root = fixture(tmp_path)
    payloads = {"spec.json": b'{"fixture": true}', "image.png": b"\x89PNG\x00fixture"}
    for name, payload in payloads.items():
        (root / name).write_bytes(payload)
        (root / f"{name}.license").write_text("SPDX-License-Identifier: Apache-2.0\n")
    assert licensing.check(root, list(payloads)) == []
    for name, payload in payloads.items():
        assert (root / name).read_bytes() == payload


def test_closest_notice_and_specific_last_annotation_win(tmp_path):
    root = fixture(
        tmp_path,
        """
[[annotations]]
path = "vendor/known-v1/**"
SPDX-License-Identifier = "Apache-2.0"
[[annotations]]
path = ["vendor/known-v1/minio/**"]
SPDX-License-Identifier = "AGPL-3.0-or-later"
""",
    )
    files = {
        "vendor/known-v1/source": "fixture",
        "vendor/known-v1/minio/source": "fixture",
        "vendor/known-v1/minio/upstream": "# SPDX-License-Identifier: MIT",
        "vendor/unknown-v2/source": "fixture",
    }
    for name, content in files.items():
        file = root / name
        file.parent.mkdir(parents=True, exist_ok=True)
        file.write_text(content)
    entries = licensing.annotations(root)
    assert licensing.identifier(root, "vendor/known-v1/source", entries) == "Apache-2.0"
    assert (
        licensing.identifier(root, "vendor/known-v1/minio/source", entries) == "AGPL-3.0-or-later"
    )
    assert licensing.identifier(root, "vendor/known-v1/minio/upstream", entries) == "MIT"
    assert licensing.check(root, list(files)) == ["vendor/unknown-v2/source"]


def test_companion_precedes_existing_header(tmp_path):
    root = fixture(tmp_path)
    (root / "source").write_text("# SPDX-License-Identifier: Apache-2.0")
    (root / "source.license").write_text("SPDX-License-Identifier: MIT")
    assert licensing.check(root, ["source"]) == []
    assert licensing.identifier(root, "source", []) == "MIT"


def test_symlink_target_is_not_read(tmp_path):
    root = fixture(tmp_path)
    link = root / "linked"
    link.symlink_to("missing-target")
    assert licensing.header_identifier(link) is None
    assert licensing.check(root, ["linked"]) == ["linked"]


@pytest.mark.parametrize(
    "metadata",
    [
        "version = 2\n",
        'version = 1\n[[annotations]]\npath = "source"\nprecedence = "override"\n',
        "version = 1\n[[annotations]]\npath = []\n",
        "version = 1\n[[annotations]]\npath = [1]\n",
        'version = 1\n[[annotations]]\npath = "../outside"\n',
        'version = 1\n[[annotations]]\npath = "/outside"\n',
        'version = 1\n[[annotations]]\npath = "*.py"\n',
    ],
)
def test_metadata_cannot_override_or_escape_notices(tmp_path, metadata):
    (tmp_path / "REUSE.toml").write_text(metadata)
    with pytest.raises(ValueError):
        licensing.annotations(tmp_path)


def test_license_texts_are_identified_by_canonical_filename(tmp_path):
    root = fixture(tmp_path)
    assert licensing.check(root, ["LICENSES/MIT.txt"]) == []


def test_cli_reports_only_paths_and_fails_missing_declarations(tmp_path, monkeypatch, capsys):
    root = fixture(tmp_path)
    (root / "source").write_text("private fixture payload; must not appear in output")
    monkeypatch.setattr(licensing, "ROOT", root)
    monkeypatch.setattr(licensing.subprocess, "check_output", lambda *args, **kwargs: b"source\0")
    assert licensing.main() == 1
    output = capsys.readouterr()
    assert "source" in output.err
    assert "private fixture payload" not in output.err
    (root / "source.license").write_text("SPDX-License-Identifier: MIT")
    assert licensing.main() == 0
    assert "1 tracked files" in capsys.readouterr().out
