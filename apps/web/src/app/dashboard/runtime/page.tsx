"use client";

import { useEffect, useState } from "react";
import { Loader2 } from "lucide-react";

export default function GenericAPIPage() {
	const [data, setData] = useState<any>(null);
	const [loading, setLoading] = useState(true);
	const [error, setError] = useState("");
	const [path, setPath] = useState("");

	useEffect(() => {
		setPath(window.location.pathname);
	}, []);

	const apiPath = path ? path.replace("/dashboard/", "/api/") : "";
	const pageName = path
		? path.split("/").pop()?.replace(/-/g, " ") || "Page"
		: "";

	useEffect(() => {
		if (!apiPath) return;
		fetch(`/api/go${apiPath}`)
			.then((r) => r.json().catch(() => ({ raw: true })))
			.then((d) => {
				setData(d);
				setLoading(false);
			})
			.catch((e) => {
				setError(String(e));
				setLoading(false);
			});
	}, [apiPath]);

	return (
		<div className="p-6 space-y-4">
			<h1 className="text-2xl font-bold capitalize">{pageName}</h1>
			{loading && <Loader2 className="w-6 h-6 animate-spin text-zinc-500" />}
			{error && <div className="text-red-400">{error}</div>}
			{data && (
				<pre className="text-sm text-zinc-300 bg-zinc-900 rounded-lg p-4 overflow-auto max-h-[80vh]">
					{JSON.stringify(data, null, 2)}
				</pre>
			)}
		</div>
	);
}
