#!/usr/bin/env python3
"""
SMLGOAPI AI Agent Integration Example
=====================================

This script demonstrates how an AI agent can use the /guide endpoint
to automatically discover and interact with the SMLGOAPI.
"""

import requests
import json
import sys
from typing import Dict, List, Any

class SMLGOAPIAgent:
    """AI Agent for interacting with SMLGOAPI"""
    
    def __init__(self, base_url: str = "http://localhost:8008"):
        self.base_url = base_url
        self.guide_data = None
        self.endpoints = None
        
    def discover_api(self) -> bool:
        """
        Discover API capabilities using the /guide endpoint
        Returns True if successful, False otherwise
        """
        try:
            print("🔍 Discovering API capabilities...")
            response = requests.get(f"{self.base_url}/guide")
            response.raise_for_status()
            
            self.guide_data = response.json()
            self.endpoints = self.guide_data.get('endpoints', {})
            
            print(f"✅ API Discovery successful!")
            print(f"   API Name: {self.guide_data.get('api_name')}")
            print(f"   Version: {self.guide_data.get('version')}")
            print(f"   Available Endpoints: {list(self.endpoints.keys())}")
            
            return True
            
        except Exception as e:
            print(f"❌ API Discovery failed: {e}")
            return False
    
    def check_health(self) -> bool:
        """Check API health status"""
        try:
            print("🏥 Checking API health...")
            response = requests.get(f"{self.base_url}/health")
            response.raise_for_status()
            
            health_data = response.json()
            print(f"✅ API Health: {health_data.get('status')}")
            print(f"   Database: {health_data.get('database')}")
            print(f"   Version: {health_data.get('version')}")
            
            return health_data.get('status') == 'healthy'
            
        except Exception as e:
            print(f"❌ Health check failed: {e}")
            return False
    
    def execute_command(self, sql_command: str) -> Dict[str, Any]:
        """Execute SQL command using /command endpoint"""
        try:
            print(f"💻 Executing command: {sql_command[:50]}...")
            
            payload = {"query": sql_command}
            response = requests.post(
                f"{self.base_url}/command",
                json=payload,
                headers={"Content-Type": "application/json"}
            )
            response.raise_for_status()
            
            result = response.json()
            if result.get('success'):
                print(f"✅ Command executed successfully in {result.get('duration_ms', 0):.2f}ms")
                return result
            else:
                print(f"❌ Command failed: {result.get('error')}")
                return result
                
        except Exception as e:
            print(f"❌ Command execution failed: {e}")
            return {"success": False, "error": str(e)}
    
    def execute_select(self, sql_query: str) -> Dict[str, Any]:
        """Execute SELECT query using /select endpoint"""
        try:
            print(f"🔍 Executing query: {sql_query[:50]}...")
            
            payload = {"query": sql_query}
            response = requests.post(
                f"{self.base_url}/select",
                json=payload,
                headers={"Content-Type": "application/json"}
            )
            response.raise_for_status()
            
            result = response.json()
            if result.get('success'):
                row_count = result.get('row_count', 0)
                duration = result.get('duration_ms', 0)
                print(f"✅ Query executed successfully: {row_count} rows in {duration:.2f}ms")
                return result
            else:
                print(f"❌ Query failed: {result.get('error')}")
                return result
                
        except Exception as e:
            print(f"❌ Query execution failed: {e}")
            return {"success": False, "error": str(e)}
    
    def search_products(self, search_term: str, limit: int = 10) -> Dict[str, Any]:
        """Search products using /search endpoint"""
        try:
            print(f"🔎 Searching products: '{search_term}' (limit: {limit})")
            
            payload = {"query": search_term, "limit": limit}
            response = requests.post(
                f"{self.base_url}/search",
                json=payload,
                headers={"Content-Type": "application/json"}
            )
            response.raise_for_status()
            
            result = response.json()
            if result.get('success'):
                total_found = result.get('metadata', {}).get('total_found', 0)
                duration = result.get('metadata', {}).get('duration_ms', 0)
                print(f"✅ Search completed: {total_found} products found in {duration:.2f}ms")
                return result
            else:
                print(f"❌ Search failed: {result.get('error')}")
                return result
                
        except Exception as e:
            print(f"❌ Search failed: {e}")
            return {"success": False, "error": str(e)}
    
    def get_database_tables(self) -> List[Dict[str, Any]]:
        """Get list of database tables"""
        try:
            print("📋 Getting database tables...")
            response = requests.get(f"{self.base_url}/api/tables")
            response.raise_for_status()
            
            tables = response.json()
            print(f"✅ Found {len(tables)} tables")
            return tables
            
        except Exception as e:
            print(f"❌ Failed to get tables: {e}")
            return []
    
    def demonstrate_capabilities(self):
        """Demonstrate AI agent capabilities"""
        print("🤖 SMLGOAPI AI Agent Demonstration")
        print("=" * 50)
        
        # Step 1: Discover API
        if not self.discover_api():
            print("Failed to discover API. Exiting.")
            return
        
        print()
        
        # Step 2: Health check
        if not self.check_health():
            print("API is not healthy. Continuing with limited functionality.")
        
        print()
        
        # Step 3: Get best practices from guide
        best_practices = self.guide_data.get('ai_agent_instructions', {}).get('best_practices', [])
        print("📚 AI Agent Best Practices:")
        for i, practice in enumerate(best_practices, 1):
            print(f"   {i}. {practice}")
          print()
        
        # Step 4: Discover database schema
        tables = self.get_database_tables()
        if tables and isinstance(tables, dict) and tables.get('success'):
            table_data = tables.get('data', [])
            print("📊 Available Tables:")
            for table in table_data[:5]:  # Show first 5 tables
                if isinstance(table, dict):
                    print(f"   - {table.get('name', 'Unknown')}")
                else:
                    print(f"   - {table}")
        elif tables:
            print(f"📊 Tables response: {type(tables)} - {tables}")
        else:
            print("📊 No tables found or error occurred")
        
        print()
        
        # Step 5: Execute test queries
        print("🧪 Testing SQL Capabilities:")
        
        # Test simple SELECT
        result = self.execute_select("SELECT 1 as test, 'Hello from AI Agent' as message, now() as timestamp")
        if result.get('success'):
            data = result.get('data', [])
            if data:
                print(f"   Test result: {data[0]}")
        
        print()
        
        # Test SHOW TABLES command
        result = self.execute_command("SHOW TABLES")
        if result.get('success'):
            print("   Tables command executed successfully")
        
        print()
          # Step 6: Test product search (if available)
        print("🔍 Testing Product Search:")
        search_result = self.search_products("test", limit=3)
        if search_result.get('success'):
            products = search_result.get('data', [])
            print(f"   Found {len(products)} products")
            if isinstance(products, list) and products:
                for product in products[:2]:  # Show first 2 products
                    if isinstance(product, dict):
                        name = product.get('product_name', 'Unknown')
                        code = product.get('product_code', 'N/A')
                        print(f"   - {name} (Code: {code})")
                    else:
                        print(f"   - {product}")
            else:
                print("   No products to display")
        
        print()
        print("🎉 AI Agent demonstration completed successfully!")
    
    def interactive_mode(self):
        """Interactive mode for testing queries"""
        print("🤖 SMLGOAPI AI Agent - Interactive Mode")
        print("Commands: 'health', 'tables', 'command <sql>', 'select <sql>', 'search <term>', 'quit'")
        print("=" * 70)
        
        if not self.discover_api():
            return
        
        while True:
            try:
                user_input = input("\n🤖 > ").strip()
                
                if user_input.lower() in ['quit', 'exit', 'q']:
                    print("👋 Goodbye!")
                    break
                
                elif user_input.lower() == 'health':
                    self.check_health()
                
                elif user_input.lower() == 'tables':
                    tables = self.get_database_tables()
                    for table in tables[:10]:  # Show first 10
                        print(f"   {table}")
                
                elif user_input.lower().startswith('command '):
                    sql = user_input[8:].strip()
                    result = self.execute_command(sql)
                    print(f"   Result: {result.get('message', 'No message')}")
                
                elif user_input.lower().startswith('select '):
                    sql = user_input[7:].strip()
                    result = self.execute_select(sql)
                    if result.get('success'):
                        data = result.get('data', [])
                        print(f"   Rows returned: {len(data)}")
                        for row in data[:3]:  # Show first 3 rows
                            print(f"   {row}")
                
                elif user_input.lower().startswith('search '):
                    term = user_input[7:].strip()
                    result = self.search_products(term)
                    if result.get('success'):
                        products = result.get('data', [])
                        print(f"   Products found: {len(products)}")
                        for product in products[:3]:  # Show first 3
                            print(f"   {product}")
                
                else:
                    print("❓ Unknown command. Try 'health', 'tables', 'command <sql>', 'select <sql>', 'search <term>', or 'quit'")
            
            except KeyboardInterrupt:
                print("\n👋 Goodbye!")
                break
            except Exception as e:
                print(f"❌ Error: {e}")


def main():
    """Main function"""
    agent = SMLGOAPIAgent()
    
    if len(sys.argv) > 1 and sys.argv[1] == '--interactive':
        agent.interactive_mode()
    else:
        agent.demonstrate_capabilities()


if __name__ == "__main__":
    main()
