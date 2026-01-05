package scaffold

import "fmt"

func indexTemplTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package views

templ Index(items []%s) {
	<div id="%s-list">
		<h1>%s List</h1>

		<table>
			<thead>
				<tr>
					<th>ID</th>
					<th>Name</th>
					<th>Actions</th>
				</tr>
			</thead>
			<tbody>
				for _, item := range items {
					<tr>
						<td>{ item.ID.String() }</td>
						<td>{ item.Name }</td>
						<td>
							<button
								hx-get={ fmt.Sprintf("/%s/%%s/edit", item.ID) }
								hx-target="#%s-form"
								hx-swap="innerHTML"
							>
								Edit
							</button>
							<button
								hx-delete={ fmt.Sprintf("/%s/%%s", item.ID) }
								hx-target="closest tr"
								hx-swap="outerHTML"
								hx-confirm="Are you sure?"
							>
								Delete
							</button>
						</td>
					</tr>
				}
			</tbody>
		</table>

		<button
			hx-get="/%s/new"
			hx-target="#%s-form"
			hx-swap="innerHTML"
		>
			New %s
		</button>

		<div id="%s-form"></div>
	</div>
}

type %s struct {
	ID   interface{ String() string }
	Name string
}
`, namePascal, name, namePascal, name, name, name, name, name, namePascal, name, namePascal)
}

func formTemplTemplate(name, namePascal string) string {
	return fmt.Sprintf(`package views

templ Form(item *%s, isNew bool) {
	<form
		if isNew {
			hx-post="/%s"
		} else {
			hx-put={ fmt.Sprintf("/%s/%%s", item.ID) }
		}
		hx-target="#%s-list"
		hx-swap="outerHTML"
	>
		<h2>
			if isNew {
				Create %s
			} else {
				Edit %s
			}
		</h2>

		<label>
			Name
			<input
				type="text"
				name="name"
				if item != nil {
					value={ item.Name }
				}
				required
			/>
		</label>

		<div>
			<button type="submit">Save</button>
			<button
				type="button"
				hx-get="/%s"
				hx-target="#%s-list"
				hx-swap="outerHTML"
			>
				Cancel
			</button>
		</div>
	</form>
}

type %s struct {
	ID   interface{ String() string }
	Name string
}
`, namePascal, name, name, name, namePascal, namePascal, name, name, namePascal)
}
