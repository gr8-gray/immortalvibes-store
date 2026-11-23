# 2026-10-02: A dead deploy token hidden for 3 weeks by a second deploy route

**What happened:** the `deploy-pages` job had failed on every push since about 2026-09-12 with
`Invalid access token [9109]`: the CI `CLOUDFLARE_API_TOKEN` had died in a credential rotation. The site kept
updating anyway, so nobody noticed.

**Root cause:** the Cloudflare Pages **Git integration** was a second, silent production route. It kept
deploying every push while the GitHub Actions route was broken, so `verify-live` and `rollback`, which only
run in Actions, were skipped for three weeks. Everything looked fine because something was deploying.

**Fix:** the CI token was replaced and `verify-live` went green again. Pages Git integration **production**
deploys were disabled (previews stay on), so GitHub Actions is the only production route. The same week,
PR #18 (`ddabaec`) gave the API its own deploy + smoke + image rollback and removed the dead Worker tier. It was
proven on push `50fe7c6` (PR #19): exactly one production build ran. Cloudflare still logs an all-`idle`
`github:push` record per push, and that's expected.

**Rules produced:** IV-2, STOP 1, STOP 28 (candidate).

**Lesson:** when two things can deploy production, the healthy one hides the broken one. Keep one route, and
make its failure loud.
