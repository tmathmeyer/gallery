import yaml
import os



GOQUERY_FMT = '''
func Query{type}Table(db *sql.DB, query map[string]interface{{}}) ([]{type}, error) {{
	rows, err := db.Query("select * from {type}_table" + gen_dep.CreateQuery(query))
	if err != nil {{
		return nil, err
	}}
	defer rows.Close()

	var data []{type}
	for rows.Next() {{
{fielddefs}
		err = rows.Scan({fieldrefs})
		if err != nil {{
			return nil, err
		}}
		data = append(data, {type}{{
			{fieldmap}
		}})
	}}
	return data, nil
}}'''

GOINSERT_FMT = '''
func Insert{type}Table(db *sql.DB, data {type}) {{
	stmt, tx, err := gen_dep.GetPreparedTransaction(db, "insert into {type}_table({sqlfields}) values({values})")
	if err != nil {{
		log.Fatal(err)
	}}
	defer stmt.Close()

	_, err = stmt.Exec({fields})
	if err != nil {{
		log.Fatal(err)
	}}

	tx.Commit()
}}'''

GOMODIFY_FMT = '''
func Update{type}Table(db *sql.DB, fieldname string, value interface{{}}, query map[string]interface{{}}) error {{
	full_query := "UPDATE {type}_table SET " + fieldname + "=? " + gen_dep.CreateQuery(query)
	stmt, tx, err := gen_dep.GetPreparedTransaction(db, full_query)
	if err != nil {{
		return err
	}}
	defer stmt.Close()

	_, err = stmt.Exec(value)
	if err != nil {{
		return err
	}}

	tx.Commit()
	return nil
}}'''

GOHEADER_FMT = '''
package generated

import (
	"database/sql"
	"../gen_dep"
	"log"
)
'''

OPENDB_FMT_A = '''
package generated

import (
	"database/sql"
	"os"
)

func file_exists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func OpenDatabase(filename string) (*sql.DB, error) {
	var fileExists = file_exists(filename)

	db, err := sql.Open("sqlite3", filename)

	if err != nil {
		return nil, err
	}

	if fileExists {
		return db, nil
	}

	create_tables := `
'''

OPENDB_FMT_B = '''
    `

	_, err = db.Exec(create_tables)
	if err != nil {
		return nil, err
	}

	return db, nil
}'''




def read_schema(f):
	with open(f) as f:
		return yaml.load(f.read())

def tdata_transform(tdata, accept_table, accept_column):
	for table, columns in tdata.items():
		yield table, table_transform(table, columns, accept_table, accept_column)

def table_transform(name, columns, accept_table, accept_column):
	return accept_table(name, [accept_column(name, types) for name,types in columns.items()])





def create_table_sqlite(name, columns):
	return 'CREATE TABLE %s_table(\n%s\n);\n\n' % (
		name.capitalize(),
		',\n'.join(columns))


def create_column_sqlite(name, types):
	return '\t%s\t%s' % (name.capitalize(), ' '.join(types))



def create_insert_func_golang(name, columns):
	fieldnames = []
	structrefs = []
	for i, column in enumerate(columns):
		fname, _ = column.strip().split('\t')
		if fname != 'Id':
			fieldnames.append(fname)
			structrefs.append('data.%s' % fname)

	return GOINSERT_FMT.format(
		type=name.capitalize(),
		sqlfields=','.join(fieldnames),
		values=','.join(['?'] * len(fieldnames)),
		fields=', '.join(structrefs))


def create_query_func_golang(name, columns):
	fieldmap = []
	fieldrefs = []
	fielddefs = []
	for i, column in enumerate(columns):
		fname, ftype = column.strip().split('\t')
		fieldmap.append('\t\t%s: %s,' % (fname, fname))
		fieldrefs.append('&%s' % fname)
		fielddefs.append('\t\tvar %s %s' % (fname, ftype))

	return GOQUERY_FMT.format(
		type=name.capitalize(),
		fieldmap='\n'.join(fieldmap),
		fielddefs='\n'.join(fielddefs),
		fieldrefs=','.join(fieldrefs))

def create_update_func_golang(name, columns):
	return GOMODIFY_FMT.format(type=name.capitalize())
	

def create_struct_golang(name, columns):
	struct = '\n'.join(['type %s struct {' % name.capitalize()] + columns + ['}'])
	query = create_query_func_golang(name, columns)
	insert = create_insert_func_golang(name, columns)
	update = create_update_func_golang(name, columns)
	return '\n'.join([GOHEADER_FMT, struct, query, insert, update])

def gotypes(types):
	for t in types:
		if t == 'INTEGER':
			return 'int'
		if t == 'REAL':
			return 'float64'
		if t == 'TEXT':
			return 'string'

def create_struct_field_golang(name, types):
	return '\t%s\t%s' % (name.capitalize(), gotypes(types))







if __name__ == '__main__':
	s = read_schema('schema.yaml')['tables']

	os.makedirs('generated', exist_ok=True)
	with open('generated/init.go', 'w') as f:
		f.write(OPENDB_FMT_A)
		for _, x in tdata_transform(s, create_table_sqlite, create_column_sqlite):
			f.write(x)
		f.write(OPENDB_FMT_B)
		

	for name, x in tdata_transform(s, create_struct_golang, create_struct_field_golang):
		with open('generated/%s.go' % name, 'w') as f:
			f.write(x)


	